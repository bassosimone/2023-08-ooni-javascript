package dsl

import (
	"crypto/tls"
	"crypto/x509"
	"errors"

	"github.com/quic-go/quic-go"
)

// QUICConnection is the results of a QUIC handshake.
type QUICConnection struct {
	// Address is the endpoint address we're using.
	Address string

	// Conn is the established QUIC connection.
	Conn quic.EarlyConnection

	// Domain is the domain we're using.
	Domain string

	// TLSConfig is the TLS configuration we used.
	TLSConfig *tls.Config

	// TLSNegotiatedProtocol is the result of the ALPN negotiation.
	TLSNegotiatedProtocol string

	// Trace is the trace we're using.
	Trace Trace
}

// QUICHandshakeOption is an option for configuring the QUIC handshake.
type QUICHandshakeOption func(config *quicHandshakeConfig)

// TODO(bassosimone): we should probably autogenerate the config, the functional optional
// setters, and the conversion from config to list of options.

type quicHandshakeConfig struct {
	ALPN       []string `json:"alpn,omitempty"`
	SkipVerify bool     `json:"skip_verify,omitempty"`
	SNI        string   `json:"sni,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	X509Certs  []string `json:"x509_certs,omitempty"`
}

func (c *quicHandshakeConfig) options() (options []QUICHandshakeOption) {
	if len(c.ALPN) > 0 {
		options = append(options, QUICHandshakeOptionALPN(c.ALPN...))
	}
	if c.SkipVerify {
		options = append(options, QUICHandshakeOptionSkipVerify(c.SkipVerify))
	}
	if c.SNI != "" {
		options = append(options, QUICHandshakeOptionSNI(c.SNI))
	}
	if len(c.Tags) > 0 {
		options = append(options, QUICHandshakeOptionTags(c.Tags...))
	}
	if len(c.X509Certs) > 0 {
		options = append(options, QUICHandshakeOptionX509Certs(c.X509Certs...))
	}
	return
}

// ErrInvalidCert is returned when we encounter an invalid PEM-encoded certificate.
var ErrInvalidCert = errors.New("dsl: invalid PEM-encoded certificate")

func (config *quicHandshakeConfig) TLSConfig() (*tls.Config, error) {
	// See https://github.com/ooni/probe/issues/2413 to understand
	// why we're using nil to force netxlite to use the cached default
	// Mozilla cert pool by default.
	out := &tls.Config{
		InsecureSkipVerify: config.SkipVerify,
		NextProtos:         config.ALPN,
		RootCAs:            nil,
		ServerName:         config.SNI,
	}

	if len(config.X509Certs) > 0 {
		certPool := x509.NewCertPool()
		for _, entry := range config.X509Certs {
			if !certPool.AppendCertsFromPEM([]byte(entry)) {
				return nil, ErrInvalidCert
			}
		}
		out.RootCAs = certPool
	}

	return out, nil
}

// QUICHandshakeOptionALPN configures the ALPN.
func QUICHandshakeOptionALPN(value ...string) QUICHandshakeOption {
	return func(config *quicHandshakeConfig) {
		config.ALPN = value
	}
}

// QUICHandshakeOptionSkipVerify allows to disable certificate verification.
func QUICHandshakeOptionSkipVerify(value bool) QUICHandshakeOption {
	return func(config *quicHandshakeConfig) {
		config.SkipVerify = value
	}
}

// QUICHandshakeOptionX509Certs allows to configure a custom root CA.
func QUICHandshakeOptionX509Certs(value ...string) QUICHandshakeOption {
	return func(config *quicHandshakeConfig) {
		config.X509Certs = value
	}
}

// QUICHandshakeOptionSNI allows to configure the SNI.
func QUICHandshakeOptionSNI(value string) QUICHandshakeOption {
	return func(config *quicHandshakeConfig) {
		config.SNI = value
	}
}

// QUICHandshakeOptionTags allows to configure the tags to include into the measurement.
func QUICHandshakeOptionTags(tags ...string) QUICHandshakeOption {
	return func(config *quicHandshakeConfig) {
		config.Tags = append(config.Tags, tags...)
	}
}

// ErrQUICHandshake wraps errors occurred during a QUIC handshake operation.
type ErrQUICHandshake struct {
	Err error
}

// Unwrap supports [errors.Unwrap].
func (exc *ErrQUICHandshake) Unwrap() error {
	return exc.Err
}

// Error implements error.
func (exc *ErrQUICHandshake) Error() string {
	return exc.Err.Error()
}

// IsErrQUICHandshake returns true when an error is an [ErrQUICHandshake].
func IsErrQUICHandshake(err error) bool {
	var exc *ErrQUICHandshake
	return errors.As(err, &exc)
}
