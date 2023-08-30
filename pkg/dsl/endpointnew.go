package dsl

import (
	"context"
	"encoding/json"
)

// NewEndpointOption is an option for configuring [NewEndpoint].
type NewEndpointOption func(op *newEndpointOperation)

// NewEndpointOptionDomain optionally configures the domain for the [*Endpoint].
func NewEndpointOptionDomain(domain string) NewEndpointOption {
	return func(op *newEndpointOperation) {
		op.Domain = domain
	}
}

// NewEndpoint returns a stage that constructs a single endpoint.
func NewEndpoint(endpoint string, options ...NewEndpointOption) Stage[*Void, *Endpoint] {
	op := &newEndpointOperation{
		Endpoint: endpoint,
		Domain:   "",
	}
	for _, option := range options {
		option(op)
	}
	return wrapOperation[*Void, *Endpoint](op)
}

type newEndpointOperation struct {
	Endpoint string `json:"endpoint"`
	Domain   string `json:"domain"`
}

const newEndpointStageName = "new_endpoint"

// ASTNode implements operation.
func (sx *newEndpointOperation) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: newEndpointStageName,
		Arguments: sx,
		Children:  []*SerializableASTNode{},
	}
}

type newEndpointLoader struct{}

// Load implements ASTLoaderRule.
func (*newEndpointLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var op newEndpointOperation
	if err := json.Unmarshal(node.Arguments, &op); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := wrapOperation[*Void, *Endpoint](&op)
	return &StageRunnableASTNode[*Void, *Endpoint]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*newEndpointLoader) StageName() string {
	return newEndpointStageName
}

// Run implements operation.
func (sx *newEndpointOperation) Run(ctx context.Context, rtx Runtime, input *Void) (*Endpoint, error) {
	if !ValidEndpoints(sx.Endpoint) {
		return nil, &ErrException{&ErrInvalidEndpoint{sx.Endpoint}}
	}
	output := &Endpoint{
		Address: sx.Endpoint,
		Domain:  sx.Domain,
	}
	return output, nil
}
