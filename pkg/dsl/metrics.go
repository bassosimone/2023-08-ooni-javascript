package dsl

import "sync"

// Metrics counts events occurring in a measurement pipeline.
type Metrics interface {
	// Error increments the error counter for the given operation metric.
	Error(name string)

	// Snapshot returns a snapshot of the metrics.
	Snapshot() map[string]int64

	// Success increments the success counter for the given operation metric.
	Success(name string)
}

// NullMetrics implements [Metrics] but ignores events. The zero value of
// this structure is ready to use.
type NullMetrics struct{}

// Error implements Metrics.
func (*NullMetrics) Error(name string) {
	// nothing
}

// Snapshot implements Metrics.
func (*NullMetrics) Snapshot() map[string]int64 {
	return make(map[string]int64)
}

// Success implements Metrics.
func (*NullMetrics) Success(name string) {
	// nothing
}

// defaultNullMetrics is the default [*NullMetrics] instance.
var defaultNullMetrics = &NullMetrics{}

// AccountingMetrics is a [Metrics] instance that accounts the events. The zero value
// of this struct is not ready to use; construct with [NewAccountingMetrics].
type AccountingMetrics struct {
	fail map[string]int64
	m    sync.Mutex
	ok   map[string]int64
}

// NewAccountingMetrics creates a new [*AccountingMetrics] instance.
func NewAccountingMetrics() *AccountingMetrics {
	return &AccountingMetrics{
		fail: map[string]int64{},
		m:    sync.Mutex{},
		ok:   map[string]int64{},
	}
}

// Error implements Metrics.
func (am *AccountingMetrics) Error(name string) {
	am.m.Lock()
	am.fail[name]++
	am.m.Unlock()
}

// Snapshot implements Metrics.
func (am *AccountingMetrics) Snapshot() map[string]int64 {
	out := make(map[string]int64)
	am.m.Lock()
	for key, value := range am.fail {
		out[key+"_error_count"] = value
	}
	for key, value := range am.ok {
		out[key+"_success_count"] = value
	}
	am.m.Unlock()
	return out
}

// Success implements Metrics.
func (am *AccountingMetrics) Success(name string) {
	am.m.Lock()
	am.ok[name]++
	am.m.Unlock()
}
