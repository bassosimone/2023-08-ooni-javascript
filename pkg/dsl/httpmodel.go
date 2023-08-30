package dsl

import (
	"errors"
	"net/http"

	"github.com/ooni/probe-engine/pkg/model"
)

// HTTPConnection is a connection usable by HTTP code.
type HTTPConnection struct {
	// Address is the remote address.
	Address string

	// Domain is the domain we're measuring.
	Domain string

	// Network is the underlying con network ("tcp" or "udp").
	Network string

	// Scheme is the URL scheme to use.
	Scheme string

	// TLSNegotiatedProtocol is the OPTIONAL negotiated protocol (e.g., "h3").
	TLSNegotiatedProtocol string

	// Trace is the trace we're using.
	Trace Trace

	// Transport is the HTTP transport wrapping the underlying conn.
	Transport model.HTTPTransport
}

// HTTPTransactionOption is an option for configuring an HTTP transaction.
type HTTPTransactionOption func(c *httpTransactionConfig)

// HTTPTransactionOptionAccept sets the Accept header.
func HTTPTransactionOptionAccept(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.AcceptHeader = value
	}
}

// HTTPTransactionOptionAcceptLanguage sets the Accept-Language header.
func HTTPTransactionOptionAcceptLanguage(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.AcceptLanguageHeader = value
	}
}

// HTTPTransactionOptionHost sets the Host header.
func HTTPTransactionOptionHost(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.HostHeader = value
	}
}

// HTTPTransactionOptionIncludeResponseBodySnapshot controls whether to include the
// response body snapshot in the JSON measurement; the default is false.
func HTTPTransactionOptionIncludeResponseBodySnapshot(value bool) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.IncludeResponseBodySnapshot = value
	}
}

// HTTPTransactionOptionMethod sets the method.
func HTTPTransactionOptionMethod(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.RequestMethod = value
	}
}

// HTTPTransactionOptionResponseBodySnapshotSize sets the maximum response body snapshot size.
func HTTPTransactionOptionResponseBodySnapshotSize(value int) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.ResponseBodySnapshotSize = value
	}
}

// HTTPTransactionOptionReferer sets the referer.
func HTTPTransactionOptionReferer(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.RefererHeader = value
	}
}

// HTTPTransactionOptionURLHost sets the URL host.
func HTTPTransactionOptionURLHost(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.URLHost = value
	}
}

// HTTPTransactionOptionURLPath sets the URL path.
func HTTPTransactionOptionURLPath(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.URLPath = value
	}
}

// HTTPTransactionOptionURLScheme sets the URL scheme.
func HTTPTransactionOptionURLScheme(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.URLScheme = value
	}
}

// HTTPTransactionOptionUserAgent sets the User-Agent header.
func HTTPTransactionOptionUserAgent(value string) HTTPTransactionOption {
	return func(c *httpTransactionConfig) {
		c.UserAgentHeader = value
	}
}

// TODO(bassosimone): we should probably autogenerate the config, the functional optional
// setters, and the conversion from config to list of options.

type httpTransactionConfig struct {
	// AcceptHeader is the accept header to use.
	AcceptHeader string `json:"accept_header,omitempty"`

	// AcceptLanguageHeader is the accept-language header to use.
	AcceptLanguageHeader string `json:"accept_language_header,omitempty"`

	// HostHeader is the host header to use.
	HostHeader string `json:"host_header,omitempty"`

	// IncludeResponseBodySnapshot tells the engine to include the response body snapshot
	// we have read inside the JSON measurement.
	IncludeResponseBodySnapshot bool `json:"include_response_body_snapshot,omitempty"`

	// RefererHeader is the referer header to use.
	RefererHeader string `json:"referer_header,omitempty"`

	// RequestMethod is the request method to use
	RequestMethod string `json:"request_method,omitempty"`

	// ResponseBodySnapshotSize is the size of the response body snapshot to read.
	ResponseBodySnapshotSize int `json:"response_body_snapshot_size,omitempty"`

	// URLHost is the host for the URL
	URLHost string `json:"url_host,omitempty"`

	// URLPath is the path for the URL
	URLPath string `json:"url_path,omitempty"`

	// URLScheme is the scheme for the URL
	URLScheme string `json:"url_scheme,omitempty"`

	// UserAgentHeader is the user-agent header to use.
	UserAgentHeader string `json:"user_agent_header,omitempty"`
}

func (c *httpTransactionConfig) options() (options []HTTPTransactionOption) {
	if value := c.AcceptHeader; value != "" {
		options = append(options, HTTPTransactionOptionAccept(value))
	}
	if value := c.AcceptLanguageHeader; value != "" {
		options = append(options, HTTPTransactionOptionAcceptLanguage(value))
	}
	if value := c.HostHeader; value != "" {
		options = append(options, HTTPTransactionOptionHost(value))
	}
	if value := c.IncludeResponseBodySnapshot; value {
		options = append(options, HTTPTransactionOptionIncludeResponseBodySnapshot(value))
	}
	if value := c.RefererHeader; value != "" {
		options = append(options, HTTPTransactionOptionReferer(value))
	}
	if value := c.RequestMethod; value != "" {
		options = append(options, HTTPTransactionOptionMethod(value))
	}
	if value := c.ResponseBodySnapshotSize; value > 0 {
		options = append(options, HTTPTransactionOptionResponseBodySnapshotSize(value))
	}
	if value := c.URLHost; value != "" {
		options = append(options, HTTPTransactionOptionURLHost(value))
	}
	if value := c.URLPath; value != "" {
		options = append(options, HTTPTransactionOptionURLPath(value))
	}
	if value := c.URLScheme; value != "" {
		options = append(options, HTTPTransactionOptionURLScheme(value))
	}
	if value := c.UserAgentHeader; value != "" {
		options = append(options, HTTPTransactionOptionUserAgent(value))
	}
	return
}

// HTTPResponse is the result of performing an HTTP transaction.
type HTTPResponse struct {
	// Address is the original endpoint address.
	Address string

	// Domain is the original domain.
	Domain string

	// Network is the original endpoint network.
	Network string

	// Request is the request we sent to the remote host.
	Request *http.Request

	// Response is the response.
	Response *http.Response

	// ResponseBodySnapshot is the body snapshot.
	ResponseBodySnapshot []byte
}

// ErrHTTPTransaction wraps errors occurred during an HTTP transaction operation.
type ErrHTTPTransaction struct {
	Err error
}

// Unwrap supports [errors.Unwrap].
func (exc *ErrHTTPTransaction) Unwrap() error {
	return exc.Err
}

// Error implements error.
func (exc *ErrHTTPTransaction) Error() string {
	return exc.Err.Error()
}

// IsErrHTTPTransaction returns true when an error is an [ErrHTTPTransaction].
func IsErrHTTPTransaction(err error) bool {
	var exc *ErrHTTPTransaction
	return errors.As(err, &exc)
}
