package dsl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ooni/probe-engine/pkg/measurexlite"
)

// DNSLookupGetaddrinfoOption is an option for [DNSLookupGetaddrinfo].
type DNSLookupGetaddrinfoOption func(operation *dnsLookupGetaddrinfoOperation)

// DNSLookupGetaddrinfoOptionTags allows configuring tags to include into measurements
// generated by the [DNSLookupGetaddrinfo] pipeline stage.
func DNSLookupGetaddrinfoOptionTags(tags ...string) DNSLookupGetaddrinfoOption {
	return func(operation *dnsLookupGetaddrinfoOperation) {
		operation.Tags = append(operation.Tags, tags...)
	}
}

// DNSLookupGetaddrinfo returns a stage that performs DNS lookups using getaddrinfo.
//
// This function returns an [ErrDNSLookup] if the error is a DNS lookup error. Remember to
// use the [IsErrDNSLookup] predicate when setting an experiment test keys.
func DNSLookupGetaddrinfo(options ...DNSLookupGetaddrinfoOption) Stage[string, *DNSLookupResult] {
	operation := &dnsLookupGetaddrinfoOperation{
		Tags: []string{},
	}
	for _, option := range options {
		option(operation)
	}
	return wrapOperation[string, *DNSLookupResult](operation)
}

type dnsLookupGetaddrinfoOperation struct {
	Tags []string `json:"tags,omitempty"`
}

const dnsLookupGetaddrinfoStageName = "dns_lookup_getaddrinfo"

// ASTNode implements operation.
func (op *dnsLookupGetaddrinfoOperation) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: dnsLookupGetaddrinfoStageName,
		Arguments: op,
		Children:  []*SerializableASTNode{},
	}
}

type dnsLookupGetaddrinfoLoader struct{}

// Load implements ASTLoaderRule.
func (*dnsLookupGetaddrinfoLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var op dnsLookupGetaddrinfoOperation
	if err := json.Unmarshal(node.Arguments, &op); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := wrapOperation[string, *DNSLookupResult](&op)
	return &StageRunnableASTNode[string, *DNSLookupResult]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*dnsLookupGetaddrinfoLoader) StageName() string {
	return dnsLookupGetaddrinfoStageName
}

// Run implements operation.
func (op *dnsLookupGetaddrinfoOperation) Run(ctx context.Context, rtx Runtime, domain string) (*DNSLookupResult, error) {
	// create trace
	trace := rtx.NewTrace(op.Tags...)

	// start the operation logger
	ol := measurexlite.NewOperationLogger(
		rtx.Logger(),
		"[#%d] DNSLookupGetaddrinfo domain=%s",
		trace.Index(),
		domain,
	)

	// setup
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// instantiate a resolver
	resolver := trace.NewStdlibResolver()

	// do the lookup
	addrs, err := resolver.LookupHost(ctx, domain)

	// stop the operation logger
	ol.Stop(err)

	// save observations
	rtx.SaveObservations(trace.ExtractObservations()...)

	// handle the error case
	if err != nil {
		rtx.Metrics().Error(dnsLookupGetaddrinfoStageName)
		return nil, &ErrDNSLookup{err}
	}

	// handle the successful case
	rtx.Metrics().Success(dnsLookupGetaddrinfoStageName)
	return &DNSLookupResult{Domain: domain, Addresses: addrs}, nil
}