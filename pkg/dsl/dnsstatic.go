package dsl

import (
	"context"
	"encoding/json"
)

// DNSLookupStatic returns a stage that always returns the given IP addresses.
func DNSLookupStatic(addresses ...string) Stage[string, *DNSLookupResult] {
	return wrapOperation[string, *DNSLookupResult](&dnsLookupStaticOperation{addresses})
}

type dnsLookupStaticOperation struct {
	Addresses []string `json:"addresses"`
}

const dnsLookupStaticStageName = "dns_lookup_static"

// ASTNode implements operation.
func (sx *dnsLookupStaticOperation) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: dnsLookupStaticStageName,
		Arguments: sx,
		Children:  []*SerializableASTNode{},
	}
}

type dnsLookupStaticLoader struct{}

// Load implements ASTLoaderRule.
func (*dnsLookupStaticLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var op dnsLookupStaticOperation
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
func (*dnsLookupStaticLoader) StageName() string {
	return dnsLookupStaticStageName
}

// Run implements operation.
func (sx *dnsLookupStaticOperation) Run(ctx context.Context, rtx Runtime, domain string) (*DNSLookupResult, error) {
	if !ValidIPAddrs(sx.Addresses...) {
		return nil, &ErrException{&ErrInvalidAddressList{sx.Addresses}}
	}
	output := &DNSLookupResult{
		Domain:    domain,
		Addresses: sx.Addresses,
	}
	return output, nil
}
