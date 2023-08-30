package dsl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

// TLSHandshake returns a stage that performs a TLS handshake.
//
// This function returns an [ErrTLSHandshake] if the error is a TLS handshake error. Remember to
// use the [IsErrTLSHandshake] predicate when setting an experiment test keys.
func TLSHandshake(options ...TLSHandshakeOption) Stage[*TCPConnection, *TLSConnection] {
	return wrapOperation[*TCPConnection, *TLSConnection](&tlsHandshakeOperation{options})
}

type tlsHandshakeOperation struct {
	options []TLSHandshakeOption
}

const tlsHandshakeStageName = "tls_handshake"

func (op *tlsHandshakeOperation) ASTNode() *SerializableASTNode {
	var config tlsHandshakeConfig
	for _, option := range op.options {
		option(&config)
	}
	return &SerializableASTNode{
		StageName: tlsHandshakeStageName,
		Arguments: &config,
		Children:  []*SerializableASTNode{},
	}
}

type tlsHandshakeLoader struct{}

// Load implements ASTLoaderRule.
func (*tlsHandshakeLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var config tlsHandshakeConfig
	if err := json.Unmarshal(node.Arguments, &config); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	stage := TLSHandshake(config.options()...)
	return &StageRunnableASTNode[*TCPConnection, *TLSConnection]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*tlsHandshakeLoader) StageName() string {
	return tlsHandshakeStageName
}

func (op *tlsHandshakeOperation) Run(ctx context.Context, rtx Runtime, tcpConn *TCPConnection) (*TLSConnection, error) {
	// initialize config
	config := &tlsHandshakeConfig{
		ALPN: []string{"h2", "http/1.1"},
		SNI:  tcpConn.Domain,
	}
	for _, option := range op.options {
		option(config)
	}

	// obtain TLS config or return an exception
	tlsConfig, err := config.TLSConfig()
	if err != nil {
		return nil, &ErrException{err}
	}

	// start the operation logger
	ol := measurexlite.NewOperationLogger(
		rtx.Logger(),
		"[#%d] TLSHandshake with %s SNI=%s ALPN=%v",
		tcpConn.Trace.Index(),
		tcpConn.Address,
		config.SNI,
		config.ALPN,
	)

	// setup
	handshaker := tcpConn.Trace.NewTLSHandshakerStdlib()
	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// handshake
	conn, state, err := handshaker.Handshake(ctx, tcpConn.Conn, tlsConfig)

	// stop the operation logger
	ol.Stop(err)

	// save observations
	rtx.SaveObservations(tcpConn.Trace.ExtractObservations()...)

	// handle the error case
	if err != nil {
		rtx.Metrics().Error(tlsHandshakeStageName)
		return nil, &ErrTLSHandshake{err}
	}

	// make sure we close this conn
	rtx.TrackCloser(conn)

	// prepare the return value
	rtx.Metrics().Success(tlsHandshakeStageName)
	out := &TLSConnection{
		Address:               tcpConn.Address,
		Conn:                  conn.(netxlite.TLSConn), // guaranteed to work
		Domain:                tcpConn.Domain,
		TLSNegotiatedProtocol: state.NegotiatedProtocol,
		Trace:                 tcpConn.Trace,
	}
	return out, nil
}
