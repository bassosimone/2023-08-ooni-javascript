package dsl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// HTTPTransaction returns a stage that uses an HTTP connection to send an HTTP request and
// reads the response headers as well as a snapshot of the response body.
//
// This function returns an [ErrHTTPTransaction] if the error is an HTTP transaction error. Remember to
// use the [IsErrHTTPTransaction] predicate when setting an experiment test keys.
func HTTPTransaction(options ...HTTPTransactionOption) Stage[*HTTPConnection, *HTTPResponse] {
	return wrapOperation[*HTTPConnection, *HTTPResponse](&httpTransactionOperation{options})
}

type httpTransactionOperation struct {
	options []HTTPTransactionOption
}

const httpTransactionStageName = "http_transaction"

// ASTNode implements operation.
func (op *httpTransactionOperation) ASTNode() *SerializableASTNode {
	var config httpTransactionConfig
	for _, option := range op.options {
		option(&config)
	}
	return &SerializableASTNode{
		StageName: httpTransactionStageName,
		Arguments: &config,
		Children:  []*SerializableASTNode{},
	}
}

type httpTransactionLoader struct{}

// Load implements ASTLoaderRule.
func (*httpTransactionLoader) Load(loader *ASTLoader, node *LoadableASTNode) (RunnableASTNode, error) {
	var config httpTransactionConfig
	if err := json.Unmarshal(node.Arguments, &config); err != nil {
		return nil, err
	}
	if err := loader.RequireExactlyNumChildren(node, 0); err != nil {
		return nil, err
	}
	options := config.options()
	stage := HTTPTransaction(options...)
	return &StageRunnableASTNode[*HTTPConnection, *HTTPResponse]{stage}, nil
}

// StageName implements ASTLoaderRule.
func (*httpTransactionLoader) StageName() string {
	return httpTransactionStageName
}

// Run implements operation.
func (op *httpTransactionOperation) Run(ctx context.Context, rtx Runtime, conn *HTTPConnection) (*HTTPResponse, error) {
	// setup
	const timeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// create configuration
	config := &httpTransactionConfig{
		AcceptHeader:                model.HTTPHeaderAccept,
		AcceptLanguageHeader:        model.HTTPHeaderAcceptLanguage,
		HostHeader:                  conn.Domain,
		IncludeResponseBodySnapshot: false,
		RefererHeader:               "",
		RequestMethod:               "GET",
		ResponseBodySnapshotSize:    1 << 19,
		URLHost:                     conn.Domain,
		URLPath:                     "/",
		URLScheme:                   conn.Scheme,
		UserAgentHeader:             model.HTTPHeaderUserAgent,
	}
	for _, option := range op.options {
		option(config)
	}

	// create HTTP request
	req, err := op.newHTTPRequest(ctx, config)
	if err != nil {
		return nil, &ErrException{err}
	}

	// start the operation logger
	ol := measurexlite.NewOperationLogger(
		rtx.Logger(),
		"[#%d] HTTPTransaction %s %s with %s/%s host=%s snapshotSize=%d",
		conn.Trace.Index(),
		config.RequestMethod,
		req.URL.String(),
		conn.Address,
		conn.Network,
		req.Host,
		config.ResponseBodySnapshotSize,
	)

	// mediate the transaction execution via the trace, which gets a chance
	// to generate HTTP observations for this transaction
	resp, body, err := conn.Trace.HTTPTransaction(
		conn,
		config.IncludeResponseBodySnapshot,
		req,
		config.ResponseBodySnapshotSize,
	)

	// save trace-collected observations (if any)
	rtx.SaveObservations(conn.Trace.ExtractObservations()...)

	// stop the operation logger
	ol.Stop(err)

	// handle the case where we failed
	if err != nil {
		rtx.Metrics().Error(httpTransactionStageName)
		return nil, &ErrHTTPTransaction{err}
	}

	// prepare the value to return
	rtx.Metrics().Success(httpTransactionStageName)
	runtimex.Assert(resp != nil, "expected response to be non-nil here")
	output := &HTTPResponse{
		Address:              conn.Address,
		Domain:               conn.Domain,
		Network:              conn.Network,
		Request:              req,
		Response:             resp,
		ResponseBodySnapshot: body,
	}
	return output, nil
}

func (op *httpTransactionOperation) newHTTPRequest(
	ctx context.Context, config *httpTransactionConfig) (*http.Request, error) {
	URL := &url.URL{
		Scheme:      config.URLScheme,
		Opaque:      "",
		User:        nil,
		Host:        config.URLHost,
		Path:        config.URLPath,
		RawPath:     "",
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	req, err := http.NewRequestWithContext(ctx, config.RequestMethod, URL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Host = config.HostHeader

	// req.Header["Host"] is ignored by Go but we want to have it in the measurement
	// to reflect what we think has been sent as HTTP headers.
	req.Header.Set("Host", req.Host)

	if v := config.AcceptHeader; v != "" {
		req.Header.Set("Accept", v)
	}

	if v := config.AcceptLanguageHeader; v != "" {
		req.Header.Set("Accept-Language", v)
	}

	if v := config.RefererHeader; v != "" {
		req.Header.Set("Referer", v)
	}

	if v := config.UserAgentHeader; v != "" { // not setting means using Go's default
		req.Header.Set("User-Agent", v)
	}

	return req, nil
}
