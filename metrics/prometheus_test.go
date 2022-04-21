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
package metrics_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent/metrics"
)

func getComponentName(t *testing.T) string {
	t.Helper()
	return strings.ReplaceAll(strings.ToLower(t.Name()), "-", "")
}

func TestPrometheusMetricsProvider_IncEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType string
	}{
		{
			name:      "increment events",
			eventType: "regular",
		},
		{
			name:      "increment errors",
			eventType: "error",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pr := prometheus.NewRegistry()

			component := getComponentName(t)
			p := metrics.NewPrometheusMetricsProviderForRegisterer(component, pr)

			nWorkers := 3
			nEvents := 10000
			nSecondaryEvents := 3
			totalEvents := nEvents * nWorkers

			var wg sync.WaitGroup

			wg.Add(nWorkers)

			var primaryFuncToUse func()
			var secondaryFuncToUse func()

			switch tt.eventType {
			case "regular":
				primaryFuncToUse = p.IncEvents
				secondaryFuncToUse = p.IncErrors
			case "error":
				primaryFuncToUse = p.IncErrors
				secondaryFuncToUse = p.IncEvents
			}

			for i := 0; i < nWorkers; i++ {
				go func(f func()) {
					defer wg.Done()

					for j := 0; j < nEvents; j++ {
						f()
					}
				}(primaryFuncToUse)
			}

			for i := 0; i < nSecondaryEvents; i++ {
				secondaryFuncToUse()
			}

			wg.Wait()

			gatheredmetrics, err := pr.Gather()
			require.NoError(t, err)
			require.Equal(t, 2, len(gatheredmetrics), "expected 2 metrics registered")

			for _, m := range gatheredmetrics {
				var buf strings.Builder
				_, fmterr := expfmt.MetricFamilyToText(&buf, m)
				require.NoError(t, fmterr)
				str := buf.String()

				var metricToCompare string

				switch tt.eventType {
				case "regular":
					switch m.GetName() {
					case metrics.EventsTotalMetricsName:
						metricToCompare = fmt.Sprintf("%s{component=%q}.*%d", metrics.EventsTotalMetricsName, component, totalEvents)
					case metrics.ErrorsTotalMetricsName:
						metricToCompare = fmt.Sprintf("%s{component=%q}.*%d", metrics.ErrorsTotalMetricsName, component, nSecondaryEvents)
					default:
						t.Errorf("unexpected metric name: %s", m.GetName())
					}
				case "error":
					switch m.GetName() {
					case metrics.EventsTotalMetricsName:
						metricToCompare = fmt.Sprintf("%s{component=%q}.*%d", metrics.EventsTotalMetricsName, component, nSecondaryEvents)
					case metrics.ErrorsTotalMetricsName:
						metricToCompare = fmt.Sprintf("%s{component=%q}.*%d", metrics.ErrorsTotalMetricsName, component, totalEvents)
					default:
						t.Errorf("unexpected metric name: %s", m.GetName())
					}
				}

				require.Regexp(t, metricToCompare, str)
			}
		})
	}
}

func TestCantRegisterMultipleTimesToSamePrometheus(t *testing.T) {
	t.Parallel()

	component := getComponentName(t)
	metrics.NewPrometheusMetricsProvider(component)

	require.Panics(t, func() {
		metrics.NewPrometheusMetricsProvider(component)
	})
}
