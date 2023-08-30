package dsl

import (
	"net/http"

	"github.com/ooni/probe-engine/pkg/model"
)

// Trace traces measurement events and produces [Observations].
type Trace interface {
	// ExtractObservations removes and returns the observations saved so far.
	ExtractObservations() []*Observations

	// HTTPTransaction executes and measures an HTTP transaction.
	//
	// Arguments:
	//
	// - conn is the HTTP connection to use;
	//
	// - includeResponseBodySnapshot controls whether to include the response body
	// snapshot into the JSON measurement;
	//
	// - request is the HTTP request;
	//
	// - responseBodySnapshotSize controls the maximum number of bytes of the
	// body that we are willing to read (to avoid reading unbounded bodies).
	//
	// Return values:
	//
	// - resp is the HTTP response;
	//
	// - body is the HTTP response body (which MAY be empty if the response body
	// snapshot size value is zero or negative);
	//
	// - err is the error that occurred (nil on success).
	HTTPTransaction(
		conn *HTTPConnection,
		includeResponseBodySnapshot bool,
		request *http.Request,
		responseBodySnapshotSize int,
	) (
		resp *http.Response,
		body []byte,
		err error,
	)

	// Index is the unique index of this trace.
	Index() int64

	// NewDialerWithoutResolver creates a dialer not attached to any resolver.
	NewDialerWithoutResolver() model.Dialer

	// NewParallelUDPResolver creates an UDP resolver resolving A and AAAA in parallel.
	NewParallelUDPResolver(endpoint string) model.Resolver

	// NewQUICDialerWithoutResolver creates a QUIC dialer not using any resolver.
	NewQUICDialerWithoutResolver() model.QUICDialer

	// NewTLSHandshakerStdlib creates a TLS handshaker using the stdlib.
	NewTLSHandshakerStdlib() model.TLSHandshaker

	// NewStdlibResolver creates a resolver using the stdlib.
	NewStdlibResolver() model.Resolver

	// Tags returns the tags configured for the trace.
	Tags() []string
}
