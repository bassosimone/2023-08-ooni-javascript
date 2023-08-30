package dsl

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/apex/log"
	"github.com/ooni/probe-engine/pkg/netxlite/filtering"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

func TestHTTPTransaction(t *testing.T) {
	t.Run("we correctly wrap HTTP transaction errors during the round trip", func(t *testing.T) {
		// create a server that RSTs during the round trip
		srvr := filtering.NewHTTPServerCleartext(filtering.HTTPActionReset)
		defer srvr.Close()

		// create a measurement pipeline
		pipeline := Compose3(
			TCPConnect(),
			HTTPConnectionTCP(),
			HTTPTransaction(),
		)

		// create the endpoint
		endpoint := NewValue(&Endpoint{
			Address: srvr.URL().Host,
			Domain:  "www.example.com",
		})

		// perform the measurement
		rtx := NewMinimalRuntime(log.Log)
		results := pipeline.Run(context.Background(), rtx, endpoint)

		// make sure the error was wrapped
		if !IsErrHTTPTransaction(results.Error) {
			t.Fatal("not an ErrHTTPTransaction", results.Error)
		}
	})

	t.Run("we correctly wrap HTTP transaction errors when reading the body", func(t *testing.T) {
		// create a server that RSTs after the HTTP round trip
		srvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// make sure we perform the round trip
			w.WriteHeader(http.StatusOK)

			// write some body
			w.Write([]byte("0xabad1deaabad1deaabad1dea"))

			// hijack the underlying connection and send the RST
			hijacker := w.(http.Hijacker)
			conn, _ := runtimex.Try2(hijacker.Hijack())
			tcpConn := conn.(*net.TCPConn)
			tcpConn.SetLinger(0)
			tcpConn.Close()
		}))
		defer srvr.Close()

		// create a measurement pipeline
		pipeline := Compose3(
			TCPConnect(),
			HTTPConnectionTCP(),
			HTTPTransaction(),
		)

		// parse the server URL
		URL := runtimex.Try1(url.Parse(srvr.URL))

		// create the endpoint
		endpoint := NewValue(&Endpoint{
			Address: URL.Host,
			Domain:  "www.example.com",
		})

		// perform the measurement
		rtx := NewMinimalRuntime(log.Log)
		results := pipeline.Run(context.Background(), rtx, endpoint)

		// make sure the error was wrapped
		if !IsErrHTTPTransaction(results.Error) {
			t.Fatal("not an ErrHTTPTransaction", results.Error)
		}
	})
}
