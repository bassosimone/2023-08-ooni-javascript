package dsl

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
)

// MakeEndpointsForPort returns a stage that converts the results of a DNS lookup to a list
// of transport layer endpoints ready to be measured using a dedicated pipeline.
func MakeEndpointsForPort(port uint16) Stage[*DNSLookupResult, []*Endpoint] {
	return &makeEndpointsForPortStage{port}
}

type makeEndpointsForPortStage struct {
	Port uint16 `json:"port"`
}

const makeEndpointsForPortStageName = "make_endpoints_for_port"

// ASTNode implements Stage.
func (sx *makeEndpointsForPortStage) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: makeEndpointsForPortStageName,
		Arguments: sx,
		Children:  []*SerializableASTNode{},
	}
}

type makeEndpointForPortLoader struct{}

// Load implements ASTLoaderRule.
func (*makeEndpointForPortLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var stage makeEndpointsForPortStage
	if err := json.Unmarshal(node.Arguments, &stage); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	return &StageRunnableASTNode[*DNSLookupResult, []*Endpoint]{&stage}, nil
}

// StageName implements ASTLoaderRule.
func (*makeEndpointForPortLoader) StageName() string {
	return makeEndpointsForPortStageName
}

// Run implements Stage.
func (sx *makeEndpointsForPortStage) Run(ctx context.Context, rtx Runtime, input Maybe[*DNSLookupResult]) Maybe[[]*Endpoint] {
	if input.Error != nil {
		return NewError[[]*Endpoint](input.Error)
	}

	// make sure we remove duplicates
	uniq := make(map[string]bool)
	for _, addr := range input.Value.Addresses {
		uniq[addr] = true
	}

	var output []*Endpoint
	for addr := range uniq {
		output = append(output, &Endpoint{
			Address: net.JoinHostPort(addr, strconv.Itoa(int(sx.Port))),
			Domain:  input.Value.Domain})
	}
	return NewValue(output)
}
