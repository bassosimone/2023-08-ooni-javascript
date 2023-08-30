package dsl

import (
	"io"
	"net/http"
	"time"

	"github.com/ooni/probe-engine/pkg/measurexlite"
	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/netxlite"
	"github.com/ooni/probe-engine/pkg/throttling"
	"github.com/quic-go/quic-go"
)

// MeasurexliteRuntime is a [Runtime] using [measurexlite] to collect [Observations].
type MeasurexliteRuntime struct {
	// metrics contains the metrics.
	metrics Metrics

	// progress is the ProgressMeter to use.
	progress ProgressMeter

	// runtime is the MinimalRuntime we compose with.
	runtime *MinimalRuntime

	// zeroTime is the zero time for observations.
	zeroTime time.Time
}

// NewMeasurexliteRuntime creates a new [MeasurexliteRuntime].
func NewMeasurexliteRuntime(
	logger model.Logger,
	metrics Metrics,
	progress ProgressMeter,
	zeroTime time.Time,
) *MeasurexliteRuntime {
	return &MeasurexliteRuntime{
		metrics:  metrics,
		progress: progress,
		runtime:  NewMinimalRuntime(logger),
		zeroTime: zeroTime,
	}
}

var _ Runtime = &MeasurexliteRuntime{}

// Close implements Runtime.
func (r *MeasurexliteRuntime) Close() error {
	return r.runtime.Close()
}

// ProgressMeter implements Runtime.
func (r *MeasurexliteRuntime) ProgressMeter() ProgressMeter {
	return r.progress
}

// Metrics implements Runtime.
func (r *MeasurexliteRuntime) Metrics() Metrics {
	return r.metrics
}

// SaveObservations implements Runtime.
func (r *MeasurexliteRuntime) SaveObservations(observations ...*Observations) {
	r.runtime.SaveObservations(observations...)
}

// ExtractObservations removes the observations from the runtime and returns them. This method
// is safe to call from multiple goroutine contexts because locks a mutex.
func (r *MeasurexliteRuntime) ExtractObservations() []*Observations {
	return r.runtime.ExtractObservations()
}

// Logger implements Runtime.
func (r *MeasurexliteRuntime) Logger() model.Logger {
	return r.runtime.Logger()
}

// TrackCloser implements Runtime.
func (r *MeasurexliteRuntime) TrackCloser(conn io.Closer) {
	r.runtime.TrackCloser(conn)
}

// TrackQUICConn implements Runtime.
func (r *MeasurexliteRuntime) TrackQUICConn(conn quic.EarlyConnection) {
	r.runtime.TrackQUICConn(conn)
}

// NewTrace implements Runtime.
func (r *MeasurexliteRuntime) NewTrace(tags ...string) Trace {
	return &measurexliteTrace{
		runtime: r,
		trace:   measurexlite.NewTrace(r.runtime.idGenerator.Add(1), r.zeroTime, tags...),
	}
}

func (r *MeasurexliteRuntime) saveNetworkEvents(events ...*model.ArchivalNetworkEvent) {
	r.SaveObservations(&Observations{NetworkEvents: events})
}

func (r *MeasurexliteRuntime) saveHTTPRequestResults(events ...*model.ArchivalHTTPRequestResult) {
	r.SaveObservations(&Observations{Requests: events})
}

// measurexliteTrace is the [Trace] returned by [MeasurexliteRuntime.NewTrace].
type measurexliteTrace struct {
	runtime *MeasurexliteRuntime
	trace   *measurexlite.Trace
}

var _ Trace = &measurexliteTrace{}

// HTTPTransaction implements Trace.
func (t *measurexliteTrace) HTTPTransaction(
	conn *HTTPConnection,
	includeResponseBodySnapshot bool,
	req *http.Request,
	maxBodySnapshotSize int,
) (*http.Response, []byte, error) {
	// make sure the response body snapshot size is non-negative
	if maxBodySnapshotSize < 0 {
		maxBodySnapshotSize = 0
	}

	tags := conn.Trace.Tags()

	// create the beginning-of-transaction observation
	started := t.trace.TimeSince(t.trace.ZeroTime)
	t.runtime.saveNetworkEvents(measurexlite.NewAnnotationArchivalNetworkEvent(
		t.trace.Index,
		started,
		"http_transaction_start",
		tags...,
	))

	// create speed sampler
	sampler := throttling.NewSampler(t.trace)
	defer sampler.Close()

	// make sure we'll know the body later on
	var body []byte

	// perform round trip
	resp, err := conn.Transport.RoundTrip(req)
	if err == nil {
		// make sure we eventually close the response body (note that closing
		// at the end of this function with `defer` would prevent the caller from
		// continuing to read the body, which isn't optimal...)
		t.runtime.TrackCloser(resp.Body)

		// TODO(bassosimone): here we should use StreamAllContext such that we
		// get a body snapshot even when we timeout reading

		// read a response-body snapshot
		reader := io.LimitReader(resp.Body, int64(maxBodySnapshotSize))
		body, err = netxlite.ReadAllContext(req.Context(), reader)
	}

	// save download speed samples before recording the end of the transaction
	// so that the last sample falls within the transaction boundaries
	t.runtime.saveNetworkEvents(sampler.ExtractSamples()...)

	// record the finish time
	finished := t.trace.TimeSince(t.trace.ZeroTime)

	// save additional network observations collected using the trace, which is
	// mainly going to be I/O events necessary to measure throttling
	t.runtime.saveNetworkEvents(t.trace.NetworkEvents()...)

	// TODO(bassosimone): when we completely omit the body, we should also
	// declare that the body has been truncated, otherwise it becomes a bit
	// difficult to understand what has actually happened. The best course
	// of action here is probably to modify
	//
	//	measurexlite.NewArchivalHTTPRequestResult
	//
	// to replace the maxBodySnapshotSize with a boolean value telling
	// the archival function whether the body has been truncated.

	// create and save an HTTP observation
	t.runtime.saveHTTPRequestResults(measurexlite.NewArchivalHTTPRequestResult(
		t.trace.Index,
		started,
		conn.Network,
		conn.Address,
		conn.TLSNegotiatedProtocol,
		conn.Transport.Network(),
		req,
		resp,
		int64(maxBodySnapshotSize),
		t.maybeIncludeBody(includeResponseBodySnapshot, body),
		err,
		finished,
		tags...,
	))

	// record that the transaction is done
	t.runtime.saveNetworkEvents(measurexlite.NewAnnotationArchivalNetworkEvent(
		t.trace.Index,
		finished,
		"http_transaction_done",
		tags...,
	))

	return resp, body, err
}

func (t *measurexliteTrace) maybeIncludeBody(includeBody bool, body []byte) []byte {
	if !includeBody {
		return []byte{}
	}
	return body
}

// Index implements Trace.
func (t *measurexliteTrace) Index() int64 {
	return t.trace.Index
}

// NewDialerWithoutResolver implements Trace.
func (t *measurexliteTrace) NewDialerWithoutResolver() model.Dialer {
	return t.trace.NewDialerWithoutResolver(t.runtime.Logger())
}

// NewParallelUDPResolver implements Trace.
func (t *measurexliteTrace) NewParallelUDPResolver(endpoint string) model.Resolver {
	return t.trace.NewParallelUDPResolver(
		t.runtime.Logger(),
		t.trace.NewDialerWithoutResolver(t.runtime.Logger()),
		endpoint,
	)
}

// NewQUICDialerWithoutResolver implements Trace.
func (t *measurexliteTrace) NewQUICDialerWithoutResolver() model.QUICDialer {
	return t.trace.NewQUICDialerWithoutResolver(netxlite.NewQUICListener(), t.runtime.Logger())
}

// NewStdlibResolver implements Trace.
func (t *measurexliteTrace) NewStdlibResolver() model.Resolver {
	return t.trace.NewStdlibResolver(t.runtime.Logger())
}

// NewTLSHandshakerStdlib implements Trace.
func (t *measurexliteTrace) NewTLSHandshakerStdlib() model.TLSHandshaker {
	return t.trace.NewTLSHandshakerStdlib(t.runtime.Logger())
}

// ExtractObservations implements Trace.
func (t *measurexliteTrace) ExtractObservations() []*Observations {
	observations := &Observations{
		NetworkEvents:  t.trace.NetworkEvents(),
		Queries:        t.trace.DNSLookupsFromRoundTrip(),
		Requests:       []*model.ArchivalHTTPRequestResult{}, // no extractor inside trace!
		TCPConnect:     t.trace.TCPConnects(),
		TLSHandshakes:  t.trace.TLSHandshakes(),
		QUICHandshakes: t.trace.QUICHandshakes(),
	}
	return []*Observations{observations}
}

// Tags implements Trace.
func (t *measurexliteTrace) Tags() []string {
	return t.trace.Tags()
}
