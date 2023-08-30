package dsl

import "github.com/ooni/probe-engine/pkg/model"

// Observations contains measurement results grouped by type.
type Observations struct {
	// NetworkEvents contains I/O events.
	NetworkEvents []*model.ArchivalNetworkEvent `json:"network_events"`

	// Queries contains the DNS queries results.
	Queries []*model.ArchivalDNSLookupResult `json:"queries"`

	// Requests contains HTTP request results.
	Requests []*model.ArchivalHTTPRequestResult `json:"requests"`

	// TCPConnect contains the TCP connect results.
	TCPConnect []*model.ArchivalTCPConnectResult `json:"tcp_connect"`

	// TLSHandshakes contains the TLS handshakes results.
	TLSHandshakes []*model.ArchivalTLSOrQUICHandshakeResult `json:"tls_handshakes"`

	// QUICHandshakes contains the QUIC handshakes results.
	QUICHandshakes []*model.ArchivalTLSOrQUICHandshakeResult `json:"quic_handshakes"`
}

// NewObservations creates an empty set of [Observations].
func NewObservations() *Observations {
	return &Observations{
		NetworkEvents:  []*model.ArchivalNetworkEvent{},
		Queries:        []*model.ArchivalDNSLookupResult{},
		Requests:       []*model.ArchivalHTTPRequestResult{},
		TCPConnect:     []*model.ArchivalTCPConnectResult{},
		TLSHandshakes:  []*model.ArchivalTLSOrQUICHandshakeResult{},
		QUICHandshakes: []*model.ArchivalTLSOrQUICHandshakeResult{},
	}
}

// ReduceObservations reduces a list of observations to a single [Observations].
func ReduceObservations(inputs ...*Observations) (output *Observations) {
	output = NewObservations()
	for _, input := range inputs {
		output.NetworkEvents = append(output.NetworkEvents, input.NetworkEvents...)
		output.QUICHandshakes = append(output.QUICHandshakes, input.QUICHandshakes...)
		output.Queries = append(output.Queries, input.Queries...)
		output.Requests = append(output.Requests, input.Requests...)
		output.TCPConnect = append(output.TCPConnect, input.TCPConnect...)
		output.TLSHandshakes = append(output.TLSHandshakes, input.TLSHandshakes...)
	}
	// TODO: we should also sort by T0 probably? or by transaction?
	return
}

// AsMap returns a map from string to any containing the observations.
func (obs *Observations) AsMap() map[string]any {
	return map[string]any{
		"network_events":  obs.NetworkEvents,
		"queries":         obs.Queries,
		"requests":        obs.Requests,
		"tcp_connect":     obs.TCPConnect,
		"tls_handshakes":  obs.TLSHandshakes,
		"quic_handshakes": obs.QUICHandshakes,
	}
}
