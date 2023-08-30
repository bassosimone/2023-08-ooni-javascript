package dsl

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/netem"
	"github.com/ooni/probe-engine/pkg/netemx"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func TestDNSLookupUDP(t *testing.T) {
	t.Run("we correctly wrap DNS lookup errors", func(t *testing.T) {
		// create the topology
		topology := runtimex.Try1(netem.NewPPPTopology(
			"10.0.0.99", "10.0.0.1", log.Log, &netem.LinkConfig{}))
		defer topology.Close()

		// create DNS server with empty DNS configuration such that a DNS lookup
		// for any domain will always return NXDOMAIN
		dnsServer := runtimex.Try1(netem.NewDNSServer(
			log.Log, topology.Server, "10.0.0.1", netem.NewDNSConfig()))
		defer dnsServer.Close()

		// run function using the client stack
		netemx.WithCustomTProxy(topology.Client, func() {
			// create an UDP pipeline
			pipeline := DNSLookupUDP("10.0.0.1:53")

			// lookup using the pipeline
			input := NewValue("www.example.com")
			rtx := NewMinimalRuntime(log.Log)
			results := pipeline.Run(context.Background(), rtx, input)

			// make sure the error is of the correct type
			if !IsErrDNSLookup(results.Error) {
				t.Fatal("not an ErrDNSLookup", results.Error)
			}
		})
	})
}
