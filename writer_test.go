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
package auditevent_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent"
	"github.com/metal-toolbox/auditevent/internal/testtools"
	"github.com/metal-toolbox/auditevent/metrics"
)

func TestEventIsSuccessfullyWritten(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectedEvent *auditevent.AuditEvent
	}{
		{
			"Basic audit event",
			auditevent.NewAuditEvent(
				"UserLogin",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"username": "ozz",
				},
				"test-login-component",
			),
		},
		{
			"audit event with target",
			auditevent.NewAuditEvent(
				"UserCreate",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeApproved,
				map[string]string{
					"username": "test",
				},
				"test-iam-component",
			).WithTarget(map[string]string{
				"path":    "/user",
				"newUser": "foobar",
			}),
		},
		{
			"audit event with data",
			auditevent.NewAuditEvent(
				"InventoryList",
				auditevent.EventSource{
					Type:  "Pod",
					Value: "network-controller-0",
					Extra: map[string]any{
						"namespace": "default",
					},
				},
				// It would be fishy if a network controller
				// was trying to list inventory. So the outcome
				// is denied.
				auditevent.OutcomeDenied,
				map[string]string{
					"rack":   "top-rack-1",
					"vendor": "ACME",
				},
				"test-lister-component",
			).WithDataFromString(`{"scope":"invalid-scope"}`),
		},
		{
			"audit event with target and data",
			auditevent.NewAuditEvent(
				"GetToken",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				"Approved",
				map[string]string{
					"username":          "requestor",
					"role":              "admin",
					"impersonated-user": "ozz",
				},
				"oidc-provider-component",
			).WithTarget(map[string]string{
				"path": "/token",
			}).WithDataFromString(`{"scope":"valid-scope"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			nppath := testtools.GetNamedPipe(t)

			// writer
			go func(eventToWrite *auditevent.AuditEvent) {
				np, err := os.OpenFile(nppath, os.O_WRONLY, 0o600)
				require.NoError(t, err)
				defer np.Close()

				aew := auditevent.NewDefaultAuditEventWriter(np)
				err = aew.Write(eventToWrite)
				require.NoError(t, err)
			}(tc.expectedEvent)

			// reader
			fd, err := os.Open(nppath)
			require.NoError(t, err)

			// This will block until the writer goroutine writes the event
			rawEvent, err := io.ReadAll(fd)
			require.NoError(t, err)

			var gotEvent auditevent.AuditEvent
			umerr := json.Unmarshal(rawEvent, &gotEvent)
			require.NoError(t, umerr)

			require.Equal(t, tc.expectedEvent.Metadata.AuditID, gotEvent.Metadata.AuditID, "audit metadata should match")
			require.Equal(t, tc.expectedEvent.Type, gotEvent.Type, "type should match")
			require.True(t, tc.expectedEvent.LoggedAt.Equal(gotEvent.LoggedAt), "logging time should match")
			require.True(t, tc.expectedEvent.LoggedAt.Before(time.Now()), "logging time should be before now")
			require.Equal(t, tc.expectedEvent.Source.Type, gotEvent.Source.Type, "source type should match")
			require.Equal(t, tc.expectedEvent.Source.Value, gotEvent.Source.Value, "source value should match")
			require.Equal(t, tc.expectedEvent.Outcome, gotEvent.Outcome, "outcome should match")
			require.Equal(t, tc.expectedEvent.Subjects, gotEvent.Subjects, "subjects should match")
			require.Equal(t, tc.expectedEvent.Component, gotEvent.Component, "component should match")
			require.Equal(t, tc.expectedEvent.Target, gotEvent.Target, "target should match")
			require.Equal(t, tc.expectedEvent.Data, gotEvent.Data, "data should match")
		})
	}
}

func TestEventMetrics(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	pr := prometheus.NewRegistry()
	w := auditevent.NewDefaultAuditEventWriter(
		&buf).WithPrometheusMetricsForRegisterer("test", pr)
	err := w.Write(
		auditevent.NewAuditEvent(
			"UserLogin",
			auditevent.EventSource{
				Type:  "IP",
				Value: "127.0.0.1",
			},
			auditevent.OutcomeSucceeded,
			map[string]string{
				"username": "ozz",
			},
			"test-login-component",
		),
	)
	require.NoError(t, err, "writing event should succeed")

	gatheredmetrics, err := pr.Gather()
	require.NoError(t, err)
	require.Equal(t, len(gatheredmetrics), 1, "should have gathered metrics")

	for _, m := range gatheredmetrics {
		var buf strings.Builder
		_, fmterr := expfmt.MetricFamilyToText(&buf, m)
		require.NoError(t, fmterr)
		str := buf.String()
		var metricToCompare string

		switch m.GetName() {
		case metrics.EventsTotalMetricsName:
			metricToCompare = fmt.Sprintf(`%s{component=%q} %d\n`, metrics.EventsTotalMetricsName, "test", 1)
		case metrics.ErrorsTotalMetricsName:
			t.Errorf("unexpected error metric name: %s", m.GetName())
		default:
			t.Errorf("unexpected metric name: %s", m.GetName())
		}

		require.Regexp(t, regexp.MustCompile(metricToCompare), str)
	}
}

func TestErrorMetrics(t *testing.T) {
	t.Parallel()

	ew := testtools.NewErrorWriter()
	pr := prometheus.NewRegistry()
	w := auditevent.NewDefaultAuditEventWriter(
		ew).WithPrometheusMetricsForRegisterer("test", pr)
	err := w.Write(
		auditevent.NewAuditEvent(
			"UserLogin",
			auditevent.EventSource{
				Type:  "IP",
				Value: "127.0.0.1",
			},
			auditevent.OutcomeSucceeded,
			map[string]string{
				"username": "ozz",
			},
			"test-login-component",
		),
	)
	require.Error(t, err, "writing event should error out")

	gatheredmetrics, err := pr.Gather()
	require.NoError(t, err)
	require.Equal(t, len(gatheredmetrics), 1, "should have gathered metrics")

	for _, m := range gatheredmetrics {
		var buf strings.Builder
		_, fmterr := expfmt.MetricFamilyToText(&buf, m)
		require.NoError(t, fmterr)
		str := buf.String()
		var metricToCompare string

		switch m.GetName() {
		case metrics.EventsTotalMetricsName:
			t.Errorf("unexpected event metric name: %s", m.GetName())
		case metrics.ErrorsTotalMetricsName:
			metricToCompare = fmt.Sprintf(`%s{component=%q} %d\n`, metrics.ErrorsTotalMetricsName, "test", 1)
		default:
			t.Errorf("unexpected metric name: %s", m.GetName())
		}

		require.Regexp(t, regexp.MustCompile(metricToCompare), str)
	}
}

func TestCantRegisterMultipleTimesToSamePrometheus(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	auditevent.NewDefaultAuditEventWriter(&buf).WithPrometheusMetrics("test")

	require.Panics(t, func() {
		auditevent.NewDefaultAuditEventWriter(&buf).WithPrometheusMetrics("test")
	})
}
