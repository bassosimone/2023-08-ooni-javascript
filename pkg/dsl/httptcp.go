package dsl

import (
	"context"

	"github.com/ooni/probe-engine/pkg/netxlite"
)

// HTTPConnectionTCP returns a stage that converts a TCP connection to an HTTP connection.
func HTTPConnectionTCP() Stage[*TCPConnection, *HTTPConnection] {
	return &httpConnectionTCPStage{}
}

type httpConnectionTCPStage struct{}

const httpConnectionTCPStageName = "http_connection_tcp"

// ASTNode implements Stage.
func (sx *httpConnectionTCPStage) ASTNode() *SerializableASTNode {
	return &SerializableASTNode{
		StageName: httpConnectionTCPStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{},
	}
}

type httpConnectionTCPLoader struct{}

// Load implements ASTLoaderRule.
func (*httpConnectionTCPLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := HTTPConnectionTCP()
	return &StageRunnableASTNode[*TCPConnection, *HTTPConnection]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*httpConnectionTCPLoader) StageName() string {
	return httpConnectionTCPStageName
}

// Run implements Stage.
func (sx *httpConnectionTCPStage) Run(ctx context.Context, rtx Runtime, input Maybe[*TCPConnection]) Maybe[*HTTPConnection] {
	if input.Error != nil {
		return NewError[*HTTPConnection](input.Error)
	}
	output := &HTTPConnection{
		Address:               input.Value.Address,
		Domain:                input.Value.Domain,
		Network:               "tcp",
		Scheme:                "http",
		TLSNegotiatedProtocol: "",
		Trace:                 input.Value.Trace,
		Transport: netxlite.NewHTTPTransport(
			rtx.Logger(), netxlite.NewSingleUseDialer(input.Value.Conn),
			netxlite.NewNullTLSDialer(),
		),
	}
	return NewValue(output)
}
