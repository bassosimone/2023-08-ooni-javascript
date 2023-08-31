package golocatev2

import (
	"context"
	"net"
	"time"

	"github.com/bassosimone/2023-08-ooni-javascript/pkg/stunx"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/pion/stun"
)

// STUNIPLookupClient discovers the probe IP using STUN. The zero value of
// this struct is invalid; please, use [NewSTUNIPLookupClient].
type STUNIPLookupClient struct {
	endpoint string
	logger   model.Logger
	resolver model.Resolver
	unet     model.UnderlyingNetwork
}

// NewSTUNIPLookupClient creates a [*STUNIPLookupClient].
func NewSTUNIPLookupClient(endpoint string,
	logger model.Logger, resolver model.Resolver, unet model.UnderlyingNetwork) *STUNIPLookupClient {
	return &STUNIPLookupClient{
		endpoint: endpoint,
		logger:   logger,
		resolver: resolver,
		unet:     unet,
	}
}

// LookupProbeIPv4 returns the IPv4 address of the probe or an error.
func (c *STUNIPLookupClient) LookupProbeIPv4(ctx context.Context) (string, error) {
	return c.lookupWithFilter(ctx, resolverFilterOnlyIPv4)
}

// LookupProbeIPv6 returns the IPv6 address of the probe or an error.
func (c *STUNIPLookupClient) LookupProbeIPv6(ctx context.Context) (string, error) {
	return c.lookupWithFilter(ctx, resolverFilterOnlyIPv6)
}

// lookupWithFilter performs the lookup applying the required filter to the addresses for the domain.
func (c *STUNIPLookupClient) lookupWithFilter(ctx context.Context, filter resolverFilter) (string, error) {
	// obtain the domainOrIPAddr (possibly an IP address) from the endpoint
	domainOrIPAddr, port, err := net.SplitHostPort(c.endpoint)
	if err != nil {
		return "", err
	}

	// create the filtered DNS resolver
	reso := newResolverWithFilter(c.resolver, filter)
	defer reso.CloseIdleConnections()

	// lookup the domain (all resolvers MUST handle IP addresses gracefully like getaddrinfo does)
	addrs, err := reso.LookupHost(ctx, domainOrIPAddr)
	if err != nil {
		return "", err
	}

	// make sure the overall operation eventually times out
	const timeout = 4 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// perform the lookup in parallel over all the given addresses
	addrch := make(chan string)
	for _, ipAddr := range addrs {
		epnt := net.JoinHostPort(ipAddr, port)
		go c.lookupWithUDPEndpoint(ctx, epnt, addrch)
	}

	// await for the first response or a timeout
	select {
	case addr := <-addrch:
		return addr, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// lookupWithUDPEndpoint performs a STUN binding-request transaction with the given endpoint.
func (c *STUNIPLookupClient) lookupWithUDPEndpoint(ctx context.Context, epnt string, addrch chan<- string) {
	// UDP connect to the remote endpoint
	netx := &netxlite.Netx{Underlying: c.unet}
	dialer := netx.NewDialerWithResolver(c.logger, &netxlite.NullResolver{})
	conn, err := dialer.DialContext(ctx, "udp", epnt)
	if err != nil {
		// TODO: what to do about this error?
		return
	}

	resp, err := stunx.RunBindingRequestTransaction(conn)
	if err != nil {
		// TODO: what to do about this error?
		return
	}

	var xorAddr stun.XORMappedAddress
	if err := xorAddr.GetFrom(resp); err != nil {
		// TODO: what to do about this error?
		return
	}
	addrch <- xorAddr.IP.String()
}
