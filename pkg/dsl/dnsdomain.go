package dsl

import (
	"context"
	"encoding/json"
)

// DomainName returns a stage that returns the given domain name.
func DomainName(value string) Stage[*Void, string] {
	return &domainNameStage{value}
}

type domainNameStage struct {
	Domain string `json:"domain"`
}

const domainNameStageName = "domain_name"

// ASTNode implements Stage.
func (sx *domainNameStage) ASTNode() *SerializableASTNode {
	// Note: we serialize the structure because this gives us forward compatibility (i.e., we
	// may add a field to a future version without breaking the AST structure and old probes will
	// be fine as long as the zero value of the new field is the default)
	return &SerializableASTNode{
		StageName: domainNameStageName,
		Arguments: sx,
		Children:  []*SerializableASTNode{},
	}
}

type domainNameLoader struct{}

// Load implements ASTLoaderRule.
func (*domainNameLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var stage domainNameStage
	if err := json.Unmarshal(node.Arguments, &stage); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	return &StageRunnableASTNode[*Void, string]{&stage}, nil
}

// StageName implements ASTLoaderRule.
func (*domainNameLoader) StageName() string {
	return domainNameStageName
}

// Run implements Stage.
func (sx *domainNameStage) Run(ctx context.Context, rtx Runtime, input Maybe[*Void]) Maybe[string] {
	if input.Error != nil {
		return NewError[string](input.Error)
	}
	if !ValidDomainNames(sx.Domain) {
		return NewError[string](&ErrException{&ErrInvalidDomain{sx.Domain}})
	}
	return NewValue(sx.Domain)
}
