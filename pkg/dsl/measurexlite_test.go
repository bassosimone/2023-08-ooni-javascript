package dsl

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/runtimex"
)

// TestMeasurexliteHTTPIncludeResponseBodySnapshot checks whether we include or not
// include a body snapshot into the JSON measurement depending on the settings.
func TestMeasurexliteHTTPIncludeResponseBodySnapshot(t *testing.T) {
	// define the expected response body
	expectedBody := []byte("Bonsoir, Elliot!\r\n")

	// create local test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(expectedBody)
	}))
	defer server.Close()

	// parse the server's URL
	URL := runtimex.Try1(url.Parse(server.URL))

	// define the function to create a measurement pipeline
	makePipeline := func(options ...HTTPTransactionOption) Stage[*Void, *HTTPResponse] {
		return Compose4(
			NewEndpoint(URL.Host, NewEndpointOptionDomain("example.com")),
			TCPConnect(),
			HTTPConnectionTCP(),
			HTTPTransaction(options...),
		)
	}

	// define the function to return the body in the pipeline
	getPipelineBody := func(input Maybe[*HTTPResponse]) ([]byte, error) {
		// handle the case of unexpected error
		if input.Error != nil {
			return nil, input.Error
		}
		return input.Value.ResponseBodySnapshot, nil
	}

	// define the function to return the body in the measurement
	getMeasurementBody := func(observations *Observations) ([]byte, error) {
		if len(observations.Requests) != 1 {
			return nil, errors.New("expected a single request entry")
		}
		return []byte(observations.Requests[0].Response.Body.Value), nil
	}

	// define the function to run the measurement
	measure := func(options ...HTTPTransactionOption) (Maybe[*HTTPResponse], *Observations) {
		pipeline := makePipeline(options...)
		meter := &NullProgressMeter{}
		rtx := NewMeasurexliteRuntime(model.DiscardLogger, &NullMetrics{}, meter, time.Now())
		input := NewValue(&Void{})
		output := pipeline.Run(context.Background(), rtx, input)
		observations := ReduceObservations(rtx.ExtractObservations()...)
		return output, observations
	}

	t.Run("the default should be that of not including the body", func(t *testing.T) {
		output, observations := measure( /* empty */ )

		t.Run("the pipeline body should contain the body", func(t *testing.T) {
			pipeBody, err := getPipelineBody(output)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("pipeline body", pipeBody)
			if diff := cmp.Diff(expectedBody, pipeBody); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("the measurement body should be empty", func(t *testing.T) {
			measBody, err := getMeasurementBody(observations)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("measurement body", measBody)
			if len(measBody) != 0 {
				t.Fatal("expected empty body")
			}
		})
	})

	t.Run("but we can optionally request to include it", func(t *testing.T) {
		output, observations := measure(HTTPTransactionOptionIncludeResponseBodySnapshot(true))

		t.Run("the pipeline body should contain the body", func(t *testing.T) {
			pipeBody, err := getPipelineBody(output)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("pipeline body", pipeBody)
			if diff := cmp.Diff(expectedBody, pipeBody); diff != "" {
				t.Fatal(diff)
			}
		})

		t.Run("the measurement body should contain the body", func(t *testing.T) {
			measBody, err := getMeasurementBody(observations)
			if err != nil {
				t.Fatal(err)
			}
			t.Log("measurement body", measBody)
			if diff := cmp.Diff(expectedBody, measBody); diff != "" {
				t.Fatal(diff)
			}
		})
	})
}
