package dsl

import (
	"fmt"
	"math"
	"net"
	"strconv"
)

// ErrInvalidDomain indicates that a domain is invalid.
type ErrInvalidDomain struct {
	Domain string
}

// Error implements error.
func (err *ErrInvalidDomain) Error() string {
	return fmt.Sprintf("dsl: invalid domain name: %s", err.Domain)
}

// ErrInvalidEndpoint indicates than an endpoint is invalid.
type ErrInvalidEndpoint struct {
	Endpoint string
}

// Error implements error.
func (err *ErrInvalidEndpoint) Error() string {
	return fmt.Sprintf("dsl: invalid endpoint: %s", err.Endpoint)
}

// ErrInvalidAddressList indicates that an address list is invalid.
type ErrInvalidAddressList struct {
	Addresses []string
}

// Error implements error.
func (err *ErrInvalidAddressList) Error() string {
	return fmt.Sprintf("dsl: invalid address list: %v", err.Addresses)
}

// ValidDomainNames returns whether the given list of domain names is valid.
func ValidDomainNames(domains ...string) bool {
	// TODO(bassosimone): how to validate domains considering IDN?
	if len(domains) <= 0 {
		return false
	}
	for _, domain := range domains {
		if len(domain) <= 0 {
			return false
		}
	}
	return true
}

// ValidIPAddrs returns whether the given list contains valid IP addresses.
func ValidIPAddrs(addrs ...string) bool {
	if len(addrs) <= 0 {
		return false
	}
	for _, addr := range addrs {
		if net.ParseIP(addr) == nil {
			return false
		}
	}
	return true
}

// ValidEndpoints returns whether the given list of endpoints is valid.
func ValidEndpoints(endpoints ...string) bool {
	if len(endpoints) <= 0 {
		return false
	}
	for _, endpoint := range endpoints {
		addr, port, err := net.SplitHostPort(endpoint)
		if err != nil {
			return false
		}
		if !ValidIPAddrs(addr) {
			return false
		}
		if !ValidPorts(port) {
			return false
		}
	}
	return true
}

// ValidPorts returns true if the given ports are valid.
func ValidPorts(ports ...string) bool {
	if len(ports) <= 0 {
		return false
	}
	for _, port := range ports {
		number, err := strconv.Atoi(port)
		if err != nil {
			return false
		}
		if number < 0 || number > math.MaxUint16 {
			return false
		}
	}
	return true
}
