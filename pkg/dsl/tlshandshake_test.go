package dsl

import (
	"context"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/netxlite/filtering"
)

func TestTLSHandshake(t *testing.T) {
	t.Run("we correctly wrap TLS handshake errors", func(t *testing.T) {
		// create a server that RSTs during the handshake
		srvr := filtering.NewTLSServer(filtering.TLSActionReset)
		defer srvr.Close()

		// create a measurement pipeline
		pipeline := Compose(
			TCPConnect(),
			TLSHandshake(),
		)

		// create the endpoint
		endpoint := NewValue(&Endpoint{
			Address: srvr.Endpoint(),
			Domain:  "www.example.com",
		})

		// perform the measurement
		rtx := NewMinimalRuntime(log.Log)
		results := pipeline.Run(context.Background(), rtx, endpoint)

		// make sure the error was wrapped
		if !IsErrTLSHandshake(results.Error) {
			t.Fatal("not an ErrTLSHandshake", results.Error)
		}
	})
}
