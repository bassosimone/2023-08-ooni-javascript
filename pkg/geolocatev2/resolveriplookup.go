package golocatev2

import (
	"context"
	"time"

	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// ResolverIPLookupClient discovers the breakout IP addresses used by the system resolver.
//
// The zero value is invalid; use [NewResolverIPLookupClient].
type ResolverIPLookupClient struct {
	// logger is the logger to use.
	logger model.Logger

	// unet is the underlying network to use.
	unet model.UnderlyingNetwork
}

// NewResolverIPLookupClient creates a new [*ResolverIPLookupClient] instance.
func NewResolverIPLookupClient(logger model.Logger, unet model.UnderlyingNetwork) *ResolverIPLookupClient {
	return &ResolverIPLookupClient{
		logger: logger,
		unet:   unet,
	}
}

// LookupResolverIPv4Addr lookups the breakout IPv4 addressed used by the system resolver.
func (ril *ResolverIPLookupClient) LookupResolverIPv4Addr(ctx context.Context) (string, error) {
	return ril.lookup(ctx, "whoami.v4.powerdns.org")
}

// LookupResolverIPv6Addr lookups the breakout IPv6 address used by the system resolver.
func (ril *ResolverIPLookupClient) LookupResolverIPv6Addr(ctx context.Context) (string, error) {
	return ril.lookup(ctx, "whoami.v6.powerdns.org")
}

func (ril *ResolverIPLookupClient) lookup(ctx context.Context, domain string) (string, error) {
	// make sure the operation is time bounded
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	// MUST be the system resolver! See https://github.com/ooni/probe/issues/2360
	addrs, _, err := ril.unet.GetaddrinfoLookupANY(ctx, domain)
	if err != nil {
		return "", err
	}

	// Note: it feels okay to panic here because a resolver is expected to never return
	// zero valid IP addresses to the caller without emitting an error.
	runtimex.Assert(len(addrs) >= 1, "reso.LookupHost returned zero IP addresses")
	return addrs[0], nil
}
