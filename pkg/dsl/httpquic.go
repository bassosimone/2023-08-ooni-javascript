package dsl

import (
	"context"

	"github.com/ooni/probe-engine/pkg/netxlite"
)

// HTTPConnectionQUIC returns a stage that converts a QUIC connection to an HTTP connection.
func HTTPConnectionQUIC() Stage[*QUICConnection, *HTTPConnection] {
	return &httpConnectionQUICStage{}
}

type httpConnectionQUICStage struct{}

const httpConnectionQUICStageName = "http_connection_quic"

// ASTNode implements Stage.
func (sx *httpConnectionQUICStage) ASTNode() *SerializableASTNode {
	return &SerializableASTNode{
		StageName: httpConnectionQUICStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{},
	}
}

type httpConnectionQUICLoader struct{}

// Load implements ASTLoaderRule.
func (*httpConnectionQUICLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := HTTPConnectionQUIC()
	return &StageRunnableASTNode[*QUICConnection, *HTTPConnection]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*httpConnectionQUICLoader) StageName() string {
	return httpConnectionQUICStageName
}

// Run implements Stage.
func (sx *httpConnectionQUICStage) Run(ctx context.Context, rtx Runtime, input Maybe[*QUICConnection]) Maybe[*HTTPConnection] {
	if input.Error != nil {
		return NewError[*HTTPConnection](input.Error)
	}
	output := &HTTPConnection{
		Address:               input.Value.Address,
		Domain:                input.Value.Domain,
		Network:               "udp",
		Scheme:                "https",
		TLSNegotiatedProtocol: input.Value.TLSNegotiatedProtocol,
		Trace:                 input.Value.Trace,
		Transport: netxlite.NewHTTP3Transport(
			rtx.Logger(), netxlite.NewSingleUseQUICDialer(input.Value.Conn),
			input.Value.TLSConfig,
		),
	}
	return NewValue(output)
}
