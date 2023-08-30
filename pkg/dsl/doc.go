// Package dsl contains a DSL for defining network experiments.
//
// # Design
//
// The DSL is both internal and external. When you write code in terms of its
// primitives, such as [TCPConnect] and [TLSHandshake], you compose a measurement
// pipeline consisting of multiple pipeline [Stage]. The composed pipeline is a
// toplevel [Stage] that runs all the underlying [Stage] performing the basic
// operations, such as [TCPConnect] and [TLSHandshake]. You can then run the composed
// pipeline by calling the [Stage] Run method and using a [Runtime] that fits your
// use case. Use a [MeasurexliteRuntime] if you need to collect [Observations] to
// create OONI measurements; use a [MinimalRuntime] otherwise.
//
// You can also serialize the measurement pipeline to JSON by converting a [Stage]
// to a [SerializableASTNode] using the [Stage] ASTNode method. In turn, you can
// obtain a JSON serialization by calling the [encoding/json.Marshal] function on
// the [SerializableASTNode]. Then, you can parse the JSON using [encoding/json.Unmarshal]
// to obtain a [LoadableASTNode]. In turn, you can transform a [LoadableASTNode] to
// a [RunnableASTNode] by using the [ASTLoader.Load] method. The [RunnableASTNode] is a
// wrapper for a [Stage] that has a generic-typing interface (e.g., uses any) and performs
// type checking at runtime. By calling a [RunnableASTNode] Run method, you run the
// generic pipeline and obtain the same results you would have obtained had you called
// the original composed pipeline [Stage] Run method.
//
// # Design Goals
//
// We designed this DSL for three reasons:
//
// 1. The internal DSL allows us to write code to generate the external DSL and,
// because the internal DSL code is also executable, we can run and test it directly
// and we can be confident that composition works because each [Stage] is bound
// to a specific input and output type and Go performs type checking.
//
// 2. The external DSL allows us to serve richer input to OONI Probes (e.g., we
// can serve measurement pipelines with specific options, including options that
// control the TLS and QUIC handshake, and we can combine multiple DNS lookup
// operations together). All in all, this functionality allows us to modify the
// implementation of simpler experiments such as Facebook Messenger using the OONI
// backend. In turn, this allows us to improve the implementation of experiments
// or fix small bugs (e.g., changes in the CA required by Signal) without releasing
// a new version of OONI Probe.
//
// 3. The external DSL also allows us to serve experimental nettests to OONI
// Probes that allow us to either perform A/B testing of nettests implementation
// or collect additional/follow-up measurements to understand censorship.
//
// Because of the above requirements, the DSL is not Turing complete. The only operation
// it offers is that of composing together network measurement primitives using
// the [Compose] function and syntactic sugar such as [Compose3], [Compose4], and
// other composition functions. In addition, it also includes specialized composition
// operations such as [DNSLookupParallel] for performing DNS lookups in parallel
// and conversion operations such as [MakeEndpointsForPort] and [NewEndpointPipeline]
// that allow to compose DNS lookups and endpoint operations. In other words, all
// you can build with the DSL is a tree that you can visit to measure the internet. There
// are no loops and there are no conditional statements.
//
// # Writing filters
//
// We additionally include functionality to register filtering functions in the
// implementation of experiments, to compute the test keys. The key feature enabling
// us to register filters is [ASTLoader.RegisterCustomLoaderRule]. Also, the
// [IfFilterExists] allows ignoring filters that a OONI Probe does not support
// during [ASTLoader.Load]. This functionality allows us to serve to probes ASTs
// including new filters that only new probes support. Older probes will honor
// the [IfFilterExists] [Stage] and replace unknow filters with the [Identity] [Stage]. This
// means that older probe would not compute some new top-level test keys but would
// otherwise collect the same [Observations] collected by new probes.
//
// When writing filters you MUST remember the following:
//
// 1. [ErrException] indicates that something went very wrong (you can test for this
// class of errors using the [IsErrException] predicate);
//
// 2. [ErrSkip] indicates that a previous stage determined that subsequent stages should
// not run (you can use [IsErrSkip] for this class of errors);
//
// 3. [ErrDNSLookup] means a DNS lookup failed (you can use [IsErrDNSLookup]);
//
// 4. [ErrTCPConnect] indicates that TCP connect failed (you can use [IsErrTCPConnect]);
//
// 5. [ErrTLSHandshake] indicates a TLS handshake failure  (use [IsErrTLSHandshake]);
//
// 6. [ErrQUICHandshake] relates to QUIC handhsake failures (use [IsErrQUICHandshake]);
//
// 7. [ErrHTTPTransaction] is an HTTP transaction error (use [IsErrHTTPTransaction]).
//
// You SHOULD only flip test keys when the error you set corresponds to the operation for
// which you are filtering errors. For example, if you filter the results of a TLS handshake,
// your code SHOULD use [IsErrTLSHandshake] to ensure that the error you are witnessing has
// indeed been caused by a TLS handshake.
//
// # History
//
// This package is an incremental evolution of the [dslx design document] where we
// added code to express the whole measurement pipeline using the DSL, rather
// than depending on imperative code to connect the DNS and the endpoints
// subpipelines. In turn, this new functionality allows us to serialize a measurement pipeline
// and serve it to the OONI Probes. The original motivation of making network experiments
// more intuitive and composable still holds.
//
// [dslx design document]: https://github.com/ooni/probe-cli/blob/master/docs/design/dd-005-dslx.md
package dsl
