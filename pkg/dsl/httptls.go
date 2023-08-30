package dsl

import (
	"context"

	"github.com/ooni/probe-engine/pkg/netxlite"
)

// HTTPConnectionTLS returns a stage that converts a TLS connection to an HTTP connection.
func HTTPConnectionTLS() Stage[*TLSConnection, *HTTPConnection] {
	return &httpConnectionTLSStage{}
}

type httpConnectionTLSStage struct{}

const httpConnectionTLSStageName = "http_connection_tls"

// ASTNode implements Stage.
func (sx *httpConnectionTLSStage) ASTNode() *SerializableASTNode {
	return &SerializableASTNode{
		StageName: httpConnectionTLSStageName,
		Arguments: nil,
		Children:  []*SerializableASTNode{},
	}
}

type httpConnectionTLSLoader struct{}

// Load implements ASTLoaderRule.
func (*httpConnectionTLSLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	if err := loader.LoadEmptyArguments(node); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := HTTPConnectionTLS()
	return &StageRunnableASTNode[*TLSConnection, *HTTPConnection]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*httpConnectionTLSLoader) StageName() string {
	return httpConnectionTLSStageName
}

// Run implements Stage.
func (sx *httpConnectionTLSStage) Run(ctx context.Context, rtx Runtime, input Maybe[*TLSConnection]) Maybe[*HTTPConnection] {
	if input.Error != nil {
		return NewError[*HTTPConnection](input.Error)
	}
	output := &HTTPConnection{
		Address:               input.Value.Address,
		Domain:                input.Value.Domain,
		Network:               "tcp",
		Scheme:                "https",
		TLSNegotiatedProtocol: input.Value.TLSNegotiatedProtocol,
		Trace:                 input.Value.Trace,
		Transport: netxlite.NewHTTPTransport(rtx.Logger(), netxlite.NewNullDialer(),
			netxlite.NewSingleUseTLSDialer(input.Value.Conn)),
	}
	return NewValue(output)
}
