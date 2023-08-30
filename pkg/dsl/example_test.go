package dsl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/bassosimone/2023-08-ooni-javascript/pkg/dsl"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// newComplexMeasurementPipeline creates a complex measurement pipeline where we
// measure two domain names in parallel: www.example.com and www.example.org.
//
// The www.example.com measurement pipeline uses three DNS resolvers in parallel and
// measures HTTP, HTTPS, and HTTP3 endpoints generated from the DNS lookup results.
//
// The www.example.org pipeline is significantly simpler. It uses getaddrinfo for DNS
// lookups and only measures the HTTPS endpoints derived from DNS lookup results.
func newComplexMeasurementPipeline() dsl.Stage[*dsl.Void, *dsl.Void] {
	// we have parallel stages for distinct target domains.
	pipeline := dsl.RunStagesInParallel(

		// this stage measures www.example.com using HTTP, HTTPS, and HTTP3
		dsl.Compose3(
			dsl.DomainName("www.example.com"),

			// we resolve the domain name using three DNS resolvers
			dsl.DNSLookupParallel(
				dsl.DNSLookupGetaddrinfo(),
				dsl.DNSLookupUDP("[2001:4860:4860::8844]:53"),
				dsl.DNSLookupUDP("8.8.4.4:53"),
			),

			// we measure the resolved IP addresses using HTTP, HTTPS, and HTTP3
			dsl.MeasureMultipleEndpoints(

				// here we measure HTTP
				dsl.Compose(
					dsl.MakeEndpointsForPort(80),
					dsl.NewEndpointPipeline(
						dsl.Compose4(
							dsl.TCPConnect(),
							dsl.HTTPConnectionTCP(),
							dsl.HTTPTransaction(),
							dsl.Discard[*dsl.HTTPResponse](),
						),
					),
				),

				// here we measure HTTPS
				dsl.Compose(
					dsl.MakeEndpointsForPort(443),
					dsl.NewEndpointPipeline(
						dsl.Compose5(
							dsl.TCPConnect(),
							dsl.TLSHandshake(),
							dsl.HTTPConnectionTLS(),
							dsl.HTTPTransaction(),
							dsl.Discard[*dsl.HTTPResponse](),
						),
					),
				),

				// here we measure HTTP3
				dsl.Compose(
					dsl.MakeEndpointsForPort(443),
					dsl.NewEndpointPipeline(
						dsl.Compose4(
							dsl.QUICHandshake(),
							dsl.HTTPConnectionQUIC(),
							dsl.HTTPTransaction(),
							dsl.Discard[*dsl.HTTPResponse](),
						),
					),
				),
			),
		),

		// this stage measures www.example.org using HTTPS
		dsl.Compose4(
			dsl.DomainName("www.example.org"),

			// we resolve the domain name using just getaddrinfo
			dsl.DNSLookupGetaddrinfo(),

			dsl.MakeEndpointsForPort(443),

			dsl.NewEndpointPipeline(
				dsl.Compose5(
					dsl.TCPConnect(),
					dsl.TLSHandshake(),
					dsl.HTTPConnectionTLS(),
					dsl.HTTPTransaction(),
					dsl.Discard[*dsl.HTTPResponse](),
				),
			),
		),
	)

	return pipeline
}

// This example shows how to use the internal DSL for measuring. We create a measurement
// pipeline using functions in the dsl package and then we run the pipeline.
func Example_internalDSL() {
	// Create the measurement pipeline.
	pipeline := newComplexMeasurementPipeline()

	// Create a measurement runtime using measurexlite as the underlying
	// measurement library such that we also collect observations.
	rtx := dsl.NewMeasurexliteRuntime(log.Log, &dsl.NullMetrics{}, &dsl.NullProgressMeter{}, time.Now())

	// Create the void input for the pipeline.
	input := dsl.NewValue(&dsl.Void{})

	// Run the measurement pipeline. The [dsl.Try] function converts the pipeline result to
	// an error if and only if an exception occurred when executing the code. We return an
	// exception when some unexpected condition occurred when measuring (e.g., the pipeline
	// defines an invalid HTTP method and we cannot create an HTTP request).
	if err := dsl.Try(pipeline.Run(context.Background(), rtx, input)); err != nil {
		log.WithError(err).Fatal("pipeline failed")
	}

	// Obtain observatins describing the performed measurement.
	observations := dsl.ReduceObservations(rtx.ExtractObservations()...)

	// Print the number of observations on the stdout.
	fmt.Printf(
		"%v %v %v %v %v %v",
		len(observations.NetworkEvents) > 0,
		len(observations.QUICHandshakes) > 0,
		len(observations.Queries) > 0,
		len(observations.Requests) > 0,
		len(observations.TCPConnect) > 0,
		len(observations.TLSHandshakes) > 0,
	)
	// output: true true true true true true
}

// This example shows how to use the internal DSL for measuring. We create a measurement
// pipeline using functions in the dsl package. Then, we serialize the pipeline to an AST,
// load the AST again, and finally execute the loaded AST.
func Example_externalDSL() {
	// Create the measurement pipeline.
	pipeline := newComplexMeasurementPipeline()

	// Serialize the measurement pipeline AST to JSON.
	rawAST := runtimex.Try1(json.Marshal(pipeline.ASTNode()))

	// Typically, you would send the serialized AST to the probe via some OONI backend API
	// such as the future check-in v2 API. In this example, we keep it simple and just pretend
	// we received the raw AST from some OONI backend API.

	// Parse the serialized JSON into an AST.
	var loadable dsl.LoadableASTNode
	runtimex.Try0(json.Unmarshal(rawAST, &loadable))

	// Create a loader for loading the AST we just parsed.
	loader := dsl.NewASTLoader()

	// Convert the AST we just loaded into a runnable AST node.
	runnable := runtimex.Try1(loader.Load(&loadable))

	// Create a measurement runtime using measurexlite as the underlying
	// measurement library such that we also collect observations.
	rtx := dsl.NewMeasurexliteRuntime(log.Log, &dsl.NullMetrics{}, &dsl.NullProgressMeter{}, time.Now())

	// Create the void input for the pipeline. We need to cast the input to
	// a generic Maybe because there's dynamic type checking when running an
	// AST we loaded from the network.
	input := dsl.NewValue(&dsl.Void{}).AsGeneric()

	// Run the measurement pipeline. The [dsl.Try] function converts the pipeline result to
	// an error if and only if an exception occurred when executing the code. We return an
	// exception when some unexpected condition occurred when measuring (e.g., the pipeline
	// defines an invalid HTTP method and we cannot create an HTTP request).
	if err := dsl.Try(runnable.Run(context.Background(), rtx, input)); err != nil {
		log.WithError(err).Fatal("pipeline failed")
	}

	// Obtain observatins describing the performed measurement.
	observations := dsl.ReduceObservations(rtx.ExtractObservations()...)

	// Print the number of observations on the stdout.
	fmt.Printf(
		"%v %v %v %v %v %v",
		len(observations.NetworkEvents) > 0,
		len(observations.QUICHandshakes) > 0,
		len(observations.Queries) > 0,
		len(observations.Requests) > 0,
		len(observations.TCPConnect) > 0,
		len(observations.TLSHandshakes) > 0,
	)
	// output: true true true true true true
}

// This example shows how to measure a single endpoint with an internal DSL.
func Example_singleEndpointInternalDSL() {
	// create a simple measurement pipeline
	pipeline := dsl.Compose4(
		dsl.NewEndpoint("8.8.8.8:443", dsl.NewEndpointOptionDomain("dns.google")),
		dsl.TCPConnect(),
		dsl.TLSHandshake(),
		dsl.Discard[*dsl.TLSConnection](),
	)

	// create the metrics
	metrics := dsl.NewAccountingMetrics()

	// create a measurement runtime
	rtx := dsl.NewMeasurexliteRuntime(log.Log, metrics, &dsl.NullProgressMeter{}, time.Now())

	// run the measurement pipeline
	_ = pipeline.Run(context.Background(), rtx, dsl.NewValue(&dsl.Void{}))

	// take a metrics snapshot
	snapshot := metrics.Snapshot()

	// print the metrics
	fmt.Printf("%+v", snapshot)

	// output: map[tcp_connect_success_count:1 tls_handshake_success_count:1]
}

// This example shows how to measure a single endpoint with an external DSL.
func Example_singleEndpointExternalDSL() {
	// create a simple measurement pipeline
	pipeline := dsl.Compose4(
		dsl.NewEndpoint("8.8.8.8:443", dsl.NewEndpointOptionDomain("dns.google")),
		dsl.TCPConnect(),
		dsl.TLSHandshake(),
		dsl.Discard[*dsl.TLSConnection](),
	)

	// Serialize the measurement pipeline AST to JSON.
	rawAST := runtimex.Try1(json.Marshal(pipeline.ASTNode()))

	// Typically, you would send the serialized AST to the probe via some OONI backend API
	// such as the future check-in v2 API. In this example, we keep it simple and just pretend
	// we received the raw AST from some OONI backend API.

	// Parse the serialized JSON into an AST.
	var loadable dsl.LoadableASTNode
	runtimex.Try0(json.Unmarshal(rawAST, &loadable))

	// Create a loader for loading the AST we just parsed.
	loader := dsl.NewASTLoader()

	// Convert the AST we just loaded into a runnable AST node.
	runnable := runtimex.Try1(loader.Load(&loadable))

	// create the metrics
	metrics := dsl.NewAccountingMetrics()

	// create a measurement runtime
	rtx := dsl.NewMeasurexliteRuntime(log.Log, metrics, &dsl.NullProgressMeter{}, time.Now())

	// run the measurement pipeline
	_ = runnable.Run(context.Background(), rtx, dsl.NewValue(&dsl.Void{}).AsGeneric())

	// take a metrics snapshot
	snapshot := metrics.Snapshot()

	// print the metrics
	fmt.Printf("%+v", snapshot)

	// output: map[tcp_connect_success_count:1 tls_handshake_success_count:1]
}
