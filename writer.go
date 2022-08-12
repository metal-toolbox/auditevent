/*
Copyright 2022 Equinix, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package auditevent

import (
	"encoding/json"
	"io"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/metal-toolbox/auditevent/metrics"
)

// EventEncoder allows for encoding audit events.
// The parameter to the `Encode` method is the audit event to encode
// and it must accept pointer to an AuditEvent struct.
type EventEncoder interface {
	Encode(any) error
}

// EventWriter writes audit events to a writer using
// a given encoder.
type EventWriter struct {
	enc EventEncoder
	mts *metrics.PrometheusMetricsProvider
}

// AuditEventEncoderJSON is an encoder that encodes audit events
// using a default JSON encoder.
func NewDefaultAuditEventWriter(w io.Writer) *EventWriter {
	enc := json.NewEncoder(w)
	return NewAuditEventWriter(enc)
}

// NewAuditEventWriter is an encoder that encodes audit events
// using the given encoder.
func NewAuditEventWriter(enc EventEncoder) *EventWriter {
	return &EventWriter{enc: enc, mts: nil}
}

// WithPrometheusMetricsForRegisterer adds prometheus metrics to this writer
// using the given prometheus registerer. It returns the writer itself for ease
// of use as the Builder pattern.
func (w *EventWriter) WithPrometheusMetricsForRegisterer(
	component string,
	pr prometheus.Registerer,
) *EventWriter {
	w.mts = metrics.NewPrometheusMetricsProviderForRegisterer(component, pr)
	return w
}

// WithPrometheusMetricsForRegisterer adds prometheus metrics to this writer
// using the default prometheus registerer (which is prometheus.DefaultRegisterer ).
// It returns the writer itself for ease of use as the Builder pattern.
func (w *EventWriter) WithPrometheusMetrics(component string) *EventWriter {
	w.mts = metrics.NewPrometheusMetricsProvider(component)
	return w
}

// Write writes an audit event to the writer.
func (w *EventWriter) Write(e *AuditEvent) error {
	err := w.enc.Encode(e)

	// We only increment the metrics if the
	// provider is available and not nil
	if w.mts != nil {
		if err == nil {
			w.mts.IncEvents()
		} else {
			w.mts.IncErrors()
		}
	}

	return err
}
