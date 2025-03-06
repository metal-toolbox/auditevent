//go:build testtools
// +build testtools

/*
Copyright 2023 Equinix, Inc.

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
// This package was tagged with testtools to ensure we don't
// pollute the library/binary with constants and functions.
package echoaudit_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent"
	"github.com/metal-toolbox/auditevent/internal/testtools"
	"github.com/metal-toolbox/auditevent/metrics"
	"github.com/metal-toolbox/auditevent/middleware/echoaudit"
)

const (
	comp = "test"
)

var testData = json.RawMessage(`{"foo":"bar"}`)

type testCase struct {
	name          string
	expectedEvent *auditevent.AuditEvent
	method        string
	headers       map[string]string
}

func getTestCases() []testCase {
	return []testCase{
		{
			"user request succeeds",
			auditevent.NewAuditEvent(
				"MyEventType",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"user": "user-ozz",
					"sub":  "sub-ozz",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/ok",
			}),
			http.MethodGet,
			nil,
		},
		{
			"user request denied with unregistered event type and unknown user",
			auditevent.NewAuditEvent(
				"GET:/denied",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeDenied,
				map[string]string{
					"user": "Unknown",
					"sub":  "Unknown",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/denied",
			}),
			http.MethodGet,
			nil,
		},
		{
			"user request denied with unregistered event type and known user",
			auditevent.NewAuditEvent(
				"GET:/denied-user",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeDenied,
				map[string]string{
					"user": "user-ozz",
					"sub":  "sub-ozz",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/denied-user",
			}),
			http.MethodGet,
			nil,
		},
		{
			"user request fails with unregistered event type and known user from header",
			auditevent.NewAuditEvent(
				"AlwaysBreaks",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeFailed,
				map[string]string{
					"user": "user-ozz-from-header",
					"sub":  "Unknown",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/fails-with-user-header",
			}),
			http.MethodPost,
			map[string]string{
				"X-User-Id": "user-ozz-from-header",
			},
		},
		{
			"user request succeeds, enriched by context data",
			auditevent.NewAuditEvent(
				"GET:/changes",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"user": "user-ozz",
					"sub":  "sub-ozz",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/changes",
			}).WithData(&testData),
			http.MethodGet,
			nil,
		},
		{
			"user request denied, encriched by context data",
			auditevent.NewAuditEvent(
				"GET:/changes/denied",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeDenied,
				map[string]string{
					"user": "Unknown",
					"sub":  "Unknown",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/changes/denied",
			}).WithData(&testData),
			http.MethodGet,
			nil,
		},
		{
			"user request succeeds, context data added as wrong type",
			auditevent.NewAuditEvent(
				"GET:/nodata",
				auditevent.EventSource{
					Type:  "IP",
					Value: "127.0.0.1",
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"user": "user-ozz",
					"sub":  "sub-ozz",
				},
				comp,
			).WithTarget(map[string]string{
				"path": "/nodata",
			}),
			http.MethodGet,
			nil,
		},
	}
}

func setFixtures(t *testing.T, w io.Writer, pr prometheus.Registerer) (*echo.Echo, *echoaudit.Middleware) {
	t.Helper()

	mdw := echoaudit.NewJSONMiddleware(comp, w)

	if pr != nil {
		mdw.WithPrometheusMetricsForRegisterer(pr)
	}

	r := echo.New()

	// All middleware are executed before the handler in echo. https://echo.labstack.com/middleware/#overview
	r.Use(mdw.AuditWithSkipper(func(c echo.Context) bool {
		return c.Request().URL.Path == "/fails-with-user-header"
	}))

	// Writing to `fails-with-user-header` breaks the app
	r.POST("/fails-with-user-header",
		func(echo.Context) error {
			return errors.New("boom") //nolint:err113 //test
		},
		mdw.AuditWithType("AlwaysBreaks"),
	)

	// allowed user with registered event type
	mdw.RegisterEventType("MyEventType", http.MethodGet, "/ok")
	r.GET("/ok", func(c echo.Context) error {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		return c.JSON(http.StatusOK, "ok")
	})

	// denied with no user
	r.GET("/denied", func(c echo.Context) error {
		return c.JSON(http.StatusForbidden, "denied")
	})

	// denied with user
	r.GET("/denied-user", func(c echo.Context) error {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		return c.JSON(http.StatusForbidden, "denied")
	})

	// allowed with user, enriched by context data
	r.GET("/changes", func(c echo.Context) error {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		c.Set(echoaudit.AuditDataContextKey, &testData)
		return c.JSON(http.StatusOK, "ok")
	})

	// denied with no user, enriched by context data
	r.GET("/changes/denied", func(c echo.Context) error {
		c.Set(echoaudit.AuditDataContextKey, &testData)
		return c.JSON(http.StatusForbidden, "denied")
	})

	// context data of wrong type
	r.GET("/nodata", func(c echo.Context) error {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		c.Set(echoaudit.AuditDataContextKey, "some random string")
		return c.JSON(http.StatusOK, "ok")
	})

	return r, mdw
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	for _, tc := range getTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := testtools.GetNamedPipe(t)

			fdchan := testtools.SetPipeReader(t, p)

			f, err := os.Open(p)
			require.NoError(t, err)

			// receive pipe reader file descriptor
			pfd := <-fdchan
			defer pfd.Close()

			r, _ := setFixtures(t, pfd, nil)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.expectedEvent.Target["path"], http.NoBody)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)

			// wait for the event to be written
			gotEvent := &auditevent.AuditEvent{}
			dec := json.NewDecoder(f)
			decErr := dec.Decode(gotEvent)
			require.NoError(t, decErr)

			require.Equal(t, tc.expectedEvent.Type, gotEvent.Type, "type should match")
			require.True(t, gotEvent.LoggedAt.Before(time.Now()), "logging time should be before now")
			require.Equal(t, tc.expectedEvent.Source.Type, gotEvent.Source.Type, "source type should match")
			require.Equal(t, tc.expectedEvent.Outcome, gotEvent.Outcome, "outcome should match")
			require.Equal(t, tc.expectedEvent.Subjects, gotEvent.Subjects, "subjects should match")
			require.Equal(t, tc.expectedEvent.Component, gotEvent.Component, "component should match")
			require.Equal(t, tc.expectedEvent.Target, gotEvent.Target, "target should match")
			require.Equal(t, tc.expectedEvent.Data, gotEvent.Data, "data should match")
			require.NotEmpty(t, gotEvent.Metadata.AuditID, "audit id is not empty")
		})
	}
}

func TestParallelCallsToMiddleware(t *testing.T) {
	t.Parallel()

	// Set up server with middleware
	p := testtools.GetNamedPipe(t)
	c := testtools.SetPipeReader(t, p)

	// set up other end of pipe
	f, err := os.OpenFile(p, os.O_RDONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	defer f.Close()

	// receive pipe reader file descriptor
	pfd := <-c

	pr := prometheus.NewRegistry()

	r, _ := setFixtures(t, pfd, pr)

	tcs := getTestCases()

	// make a bunch of requests
	// This is set to 8000 since for race testing there's a goroutine limit
	// of 8128
	nreqs := 8000
	var wg sync.WaitGroup
	for i := 0; i < nreqs; i++ {
		wg.Add(1)
		tc := tcs[i%len(tcs)]

		go func(wg *sync.WaitGroup, method, path string, headers map[string]string) {
			defer wg.Done()

			w := httptest.NewRecorder()

			req := httptest.NewRequest(method, path, http.NoBody)

			for k, v := range headers {
				req.Header.Set(k, v)
			}

			r.ServeHTTP(w, req)
		}(&wg, tc.method, tc.expectedEvent.Target["path"], tc.headers)
	}

	// close pipe when all events are received
	go func() {
		wg.Wait()
		pfd.Close()
	}()

	reader := bufio.NewReader(f)

	var numlines int
	// Read events from the pipe
	for {
		_, _, readErr := reader.ReadLine()
		if errors.Is(readErr, io.EOF) {
			break
		}
		numlines++
	}

	// verify we didn't loose audit logs
	require.Equal(t, nreqs, numlines, "number of events should match")

	gatheredmetrics, err := pr.Gather()
	require.NoError(t, err)
	require.Greater(t, len(gatheredmetrics), 0, "should have gathered metrics")

	for _, m := range gatheredmetrics {
		var buf strings.Builder
		_, fmterr := expfmt.MetricFamilyToText(&buf, m)
		require.NoError(t, fmterr)
		str := buf.String()
		var metricToCompare string

		switch m.GetName() {
		case metrics.EventsTotalMetricsName:
			metricToCompare = fmt.Sprintf(`%s{component=%q} %d\n`, metrics.EventsTotalMetricsName, comp, numlines)
		case metrics.ErrorsTotalMetricsName:
			t.Errorf("unexpected error metric name: %s", m.GetName())
		default:
			t.Errorf("unexpected metric name: %s", m.GetName())
		}

		require.Regexp(t, regexp.MustCompile(metricToCompare), str)
	}
}

func TestCantRegisterMultipleTimesToSamePrometheus(t *testing.T) {
	t.Parallel()

	var buf strings.Builder
	echoaudit.NewJSONMiddleware(comp, &buf).WithPrometheusMetrics()

	require.Panics(t, func() {
		echoaudit.NewJSONMiddleware(comp, &buf).WithPrometheusMetrics()
	})
}

// Tests that the middleware generates events with a custom outcome handler.
func TestMiddlewareWithCustomOutcomeHandler(t *testing.T) {
	t.Parallel()

	for _, tc := range getTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := testtools.GetNamedPipe(t)

			fdchan := testtools.SetPipeReader(t, p)

			f, err := os.Open(p)
			require.NoError(t, err)

			// receive pipe reader file descriptor
			pfd := <-fdchan
			defer pfd.Close()

			r, mdw := setFixtures(t, pfd, nil)
			mdw.WithOutcomeHandler(func(echo.Context) string {
				return "custom"
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.expectedEvent.Target["path"], http.NoBody)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)

			// wait for the event to be written
			gotEvent := &auditevent.AuditEvent{}
			dec := json.NewDecoder(f)
			decErr := dec.Decode(gotEvent)
			require.NoError(t, decErr)

			require.Equal(t, tc.expectedEvent.Type, gotEvent.Type, "type should match")
			require.True(t, gotEvent.LoggedAt.Before(time.Now()), "logging time should be before now")
			require.Equal(t, tc.expectedEvent.Source.Type, gotEvent.Source.Type, "source type should match")
			require.Equal(t, tc.expectedEvent.Subjects, gotEvent.Subjects, "subjects should match")
			require.Equal(t, tc.expectedEvent.Component, gotEvent.Component, "component should match")
			require.Equal(t, tc.expectedEvent.Target, gotEvent.Target, "target should match")
			require.Equal(t, tc.expectedEvent.Data, gotEvent.Data, "data should match")

			// This is the custom outcome we set above
			require.Equal(t, "custom", gotEvent.Outcome, "outcome should match")
		})
	}
}

// Tests that the middleware generates events with a custom subject handler.
func TestMiddlewareWithCustomSubjectHandler(t *testing.T) {
	t.Parallel()

	for _, tc := range getTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := testtools.GetNamedPipe(t)

			fdchan := testtools.SetPipeReader(t, p)

			f, err := os.Open(p)
			require.NoError(t, err)

			// receive pipe reader file descriptor
			pfd := <-fdchan
			defer pfd.Close()

			r, mdw := setFixtures(t, pfd, nil)
			mdw.WithSubjectHandler(func(echo.Context) map[string]string {
				return map[string]string{"custom": "customvalue"}
			})
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.expectedEvent.Target["path"], http.NoBody)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			r.ServeHTTP(w, req)

			// wait for the event to be written
			gotEvent := &auditevent.AuditEvent{}
			dec := json.NewDecoder(f)
			decErr := dec.Decode(gotEvent)
			require.NoError(t, decErr)

			require.Equal(t, tc.expectedEvent.Type, gotEvent.Type, "type should match")
			require.True(t, gotEvent.LoggedAt.Before(time.Now()), "logging time should be before now")
			require.Equal(t, tc.expectedEvent.Source.Type, gotEvent.Source.Type, "source type should match")
			require.Equal(t, tc.expectedEvent.Component, gotEvent.Component, "component should match")
			require.Equal(t, tc.expectedEvent.Target, gotEvent.Target, "target should match")
			require.Equal(t, tc.expectedEvent.Data, gotEvent.Data, "data should match")
			require.Equal(t, tc.expectedEvent.Outcome, gotEvent.Outcome, "outcome should match")

			// This is the custom subjects we set above
			require.Equal(t, map[string]string{"custom": "customvalue"}, gotEvent.Subjects, "subjects should match")
		})
	}
}
