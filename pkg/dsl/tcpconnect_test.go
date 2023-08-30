package dsl

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/netemx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func TestTCPConnect(t *testing.T) {
	t.Run("we correctly wrap TCP connect errors", func(t *testing.T) {
		// create the topology
		topology := runtimex.Try1(netem.NewPPPTopology(
			"10.0.0.99", "10.0.0.1", log.Log, &netem.LinkConfig{}))
		defer topology.Close()

		// Note: do not create any TCP listener, so the connection will fail

		// run function using the client stack
		netemx.WithCustomTProxy(topology.Client, func() {
			// create a pipeline
			pipeline := TCPConnect()

			endpoint := NewValue(&Endpoint{
				Address: "10.0.0.1:80",
				Domain:  "www.example.com",
			})

			// lookup using the pipeline
			rtx := NewMinimalRuntime(log.Log)
			results := pipeline.Run(context.Background(), rtx, endpoint)

			// make sure the error is of the correct type
			if !IsErrTCPConnect(results.Error) {
				t.Fatal("not an ErrTCPConnect", results.Error)
			}
		})
	})
}
