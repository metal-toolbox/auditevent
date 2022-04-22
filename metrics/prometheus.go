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
package metrics

import "github.com/prometheus/client_golang/prometheus"

const (
	// EventsTotalMetricsName is the name of the metric that tracks the number of events.
	EventsTotalMetricsName = "audit_events_total"

	// ErrorsTotalMetricsName is the name of the metric that tracks the number of errors
	// writing audit events.
	ErrorsTotalMetricsName = "audit_errors_total"

	// ComponentLabelName is the name of the label that identifies the component
	// This is a label used in both the "audit_events_total" and "audit_errors_total" metrics.
	ComponentLabelName = "component"
)

// PrometheusMetricsProvider is a metrics provider that uses prometheus as a backend.
type PrometheusMetricsProvider struct {
	component string
	nEvents   *prometheus.CounterVec
	nErrors   *prometheus.CounterVec
}

// NewPrometheusMetricsProviderForRegisterer returns a new instance of a metrics provider that
// uses prometheus as a backend. It requires a component name which will be used as
// a label in the metrics.
func NewPrometheusMetricsProvider(component string) *PrometheusMetricsProvider {
	return NewPrometheusMetricsProviderForRegisterer(component, prometheus.DefaultRegisterer)
}

// NewPrometheusMetricsProviderForRegisterer returns a new instance of a metrics provider that
// uses prometheus as a backend. It requires a component name which will be used as
// a label in the metrics, as well as a prometheus registry.
// Normally, the prometheus registry will be the global `prometheus.DefaultRegisterer`.
func NewPrometheusMetricsProviderForRegisterer(
	component string,
	r prometheus.Registerer,
) *PrometheusMetricsProvider {
	p := &PrometheusMetricsProvider{
		component: component,
		nEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: EventsTotalMetricsName,
				Help: "Number of audit events generated.",
			},
			[]string{ComponentLabelName},
		),
		nErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: ErrorsTotalMetricsName,
				Help: "Number of errors writing audit events.",
			},
			[]string{ComponentLabelName},
		),
	}

	for _, m := range []prometheus.Collector{p.nEvents, p.nErrors} {
		r.MustRegister(m)
	}

	return p
}

// Increase the number of audit events that have been written.
func (p *PrometheusMetricsProvider) IncEvents() {
	p.nEvents.WithLabelValues(p.component).Inc()
}

// Increase the number of audit events that have errored out.
func (p *PrometheusMetricsProvider) IncErrors() {
	p.nErrors.WithLabelValues(p.component).Inc()
}
