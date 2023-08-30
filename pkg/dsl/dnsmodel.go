package dsl

import "errors"

// DNSLookupResult is the result of a DNS lookup operation.
type DNSLookupResult struct {
	// Domain is the domain we tried to resolve.
	Domain string

	// Addresses contains resolved addresses (if any).
	Addresses []string
}

// ErrDNSLookup wraps errors occurred during a DNS lookup operation.
type ErrDNSLookup struct {
	Err error
}

// Unwrap supports [errors.Unwrap].
func (exc *ErrDNSLookup) Unwrap() error {
	return exc.Err
}

// Error implements error.
func (exc *ErrDNSLookup) Error() string {
	return exc.Err.Error()
}

// IsErrDNSLookup returns true when an error is an [ErrDNSLookup].
func IsErrDNSLookup(err error) bool {
	var exc *ErrDNSLookup
	return errors.As(err, &exc)
}
