package golocatev2

import (
	"context"
	"net"
	"strings"

	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
)

// resolverFilter returns true if the address should be accepted.
type resolverFilter func(ipAddr string) bool

// resolverFilterOnlyIPv4 returns true if the address is IPv4
func resolverFilterOnlyIPv4(ipAddr string) bool {
	return net.ParseIP(ipAddr) != nil && !strings.Contains(ipAddr, ":")
}

// resolverFilterOnlyIPv6 returns true if the address is IPv6.
func resolverFilterOnlyIPv6(ipAddr string) bool {
	return net.ParseIP(ipAddr) != nil && strings.Contains(ipAddr, ":")
}

// resolverWithFilter is a resolver using a configured [resolverFilter]. The zero value
// of this struct is invalid; please, use [newResolverWithFilter].
type resolverWithFilter struct {
	resolver model.Resolver
	filter   resolverFilter
}

// newResolverWithFilter wraps a resolver using the given filter.
func newResolverWithFilter(resolver model.Resolver, filter resolverFilter) *resolverWithFilter {
	return &resolverWithFilter{
		resolver: resolver,
		filter:   filter,
	}
}

func (rf *resolverWithFilter) LookupHost(ctx context.Context, domain string) ([]string, error) {
	orig, err := rf.resolver.LookupHost(ctx, domain)
	if err != nil {
		return nil, err
	}
	var addrs []string
	for _, addr := range orig {
		if rf.filter(addr) {
			addrs = append(addrs, addr)
		}
	}
	if len(addrs) <= 0 {
		return nil, netxlite.ErrOODNSNoAnswer
	}
	return addrs, nil
}

func (rf *resolverWithFilter) CloseIdleConnections() {
	rf.resolver.CloseIdleConnections()
}
