package dsl

import (
	"errors"
	"net"
)

// TCPConnection is the result of performing a TCP connect operation.
type TCPConnection struct {
	// Address is the endpoint address we're using.
	Address string

	// Conn is the established TCP connection.
	Conn net.Conn

	// Domain is the domain we're using.
	Domain string

	// Trace is the trace we're using.
	Trace Trace
}

// ErrTCPConnect wraps errors occurred during a TCP connect operation.
type ErrTCPConnect struct {
	Err error
}

// Unwrap supports [errors.Unwrap].
func (exc *ErrTCPConnect) Unwrap() error {
	return exc.Err
}

// Error implements error.
func (exc *ErrTCPConnect) Error() string {
	return exc.Err.Error()
}

// IsErrTCPConnect returns true when an error is an [ErrTCPConnect].
func IsErrTCPConnect(err error) bool {
	var exc *ErrTCPConnect
	return errors.As(err, &exc)
}
