package dsl

import (
	"context"
	"net"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func TestQUICHandshake(t *testing.T) {
	t.Run("we correctly wrap QUIC errors", func(t *testing.T) {
		// create a connection where we're not going to read incoming packets
		listener := netxlite.NewQUICListener()
		pconn := runtimex.Try1(listener.Listen(&net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 0, // let the kernel pick a port
			Zone: "",
		}))
		defer pconn.Close()
		localAddr := pconn.LocalAddr().String()

		// create measurement pipeline
		pipeline := QUICHandshake()

		// create the endpoint
		endpoint := NewValue(&Endpoint{
			Address: localAddr,
			Domain:  "www.example.com",
		})

		// perform the measurement
		rtx := NewMinimalRuntime(log.Log)
		results := pipeline.Run(context.Background(), rtx, endpoint)

		// make sure the error is correct
		if !IsErrQUICHandshake(results.Error) {
			t.Fatal("not an ErrQUICHandshake", results.Error)
		}
	})
}
