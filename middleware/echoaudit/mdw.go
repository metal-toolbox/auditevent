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
package echoaudit

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/metal-toolbox/auditevent"
)

const (
	// AuditDataContextKey is the context key for additional audit data.
	AuditDataContextKey = "audit.data"
	// AuditIDContextKey is the context key for the audit ID.
	AuditIDContextKey = "audit.id"
)

type Middleware struct {
	component      string
	aew            *auditevent.EventWriter
	eventTypeMap   sync.Map
	outcomeHandler OutcomeHandler
	subjectHandler SubjectHandler
}

// NewMiddleware returns a new instance of audit Middleware.
func NewMiddleware(component string, aew *auditevent.EventWriter) *Middleware {
	return &Middleware{
		component:      component,
		aew:            aew,
		outcomeHandler: GetOutcomeDefault,
		subjectHandler: GetSubjectDefault,
	}
}

// NewJSONMiddleware returns a new middleware instance with a default JSON writer.
func NewJSONMiddleware(component string, w io.Writer) *Middleware {
	return NewMiddleware(
		component,
		auditevent.NewDefaultAuditEventWriter(w),
	)
}

// WithPrometheusMetrics enables prometheus metrics for this middleware instance
// using the default prometheus registerer (prometheus.DefaultRegisterer).
func (m *Middleware) WithPrometheusMetrics() *Middleware {
	m.aew.WithPrometheusMetrics(m.component)
	return m
}

// WithPrometheusMetricsForRegisterer enables prometheus metrics for this middleware instance
// using the default prometheus registerer (prometheus.DefaultRegisterer).
func (m *Middleware) WithPrometheusMetricsForRegisterer(pr prometheus.Registerer) *Middleware {
	m.aew.WithPrometheusMetricsForRegisterer(m.component, pr)
	return m
}

func (m *Middleware) WithOutcomeHandler(handler OutcomeHandler) *Middleware {
	m.outcomeHandler = handler
	return m
}

func (m *Middleware) WithSubjectHandler(handler SubjectHandler) *Middleware {
	m.subjectHandler = handler
	return m
}

// RegisterEventType registers an audit event type for a given HTTP method and path.
func (m *Middleware) RegisterEventType(eventType, httpMethod, path string) {
	m.eventTypeMap.Store(keyFromHTTPMethodAndPath(httpMethod, path), eventType)
}

// Audit returns a echo middleware that will audit the request.
// This uses the a pre-registered type for the event (see RegisterEventType).
// If no type is registered, the event type is the HTTP method and path.
func (m *Middleware) Audit() echo.MiddlewareFunc {
	return m.AuditWithType("")
}

// MiddlewareFunc func(next HandlerFunc) HandlerFunc

// AuditWithType returns a echo middleware that will audit the request.
// This uses the given type for the event.
// If the type is empty, the event type will try to use a pre-registered
// type (see RegisterEventType) and if it doesn't find it,
// it'll use the HTTP method and path.
func (m *Middleware) AuditWithType(t string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			path := c.Request().URL.Path

			auditID := uuid.New().String()
			c.Set(AuditIDContextKey, auditID)

			// We audit after the request has been processed
			if err := next(c); err != nil {
				return err
			}

			event := auditevent.NewAuditEventWithID(
				auditID,
				m.getEventType(t, method, path),
				auditevent.EventSource{
					Type: "IP",
					// TODO we should implement logic like exists in gin for trusted proxies
					// https://github.com/gin-gonic/gin/blob/457fabd7e14f36ca1b5f302f7247efeb4690e49c/context.go#L771
					Value: c.RealIP(),
				},
				m.outcomeHandler(c),
				m.subjectHandler(c),
				m.component,
			).WithTarget(map[string]string{
				"path": path,
			})

			if data := c.Get(AuditDataContextKey); data != nil {
				ed, ok := data.(*json.RawMessage)
				if ok {
					event.WithData(ed)
				}
			}

			// persist event
			m.write(event)

			return nil
		}
	}
}

func (m *Middleware) getEventType(preferredType, httpMethod, path string) string {
	if preferredType != "" {
		return preferredType
	}

	key := keyFromHTTPMethodAndPath(httpMethod, path)
	rawEventType, ok := m.eventTypeMap.Load(key)
	if ok {
		etype, castok := rawEventType.(string)
		if castok {
			return etype
		}
	}
	return key
}

// This function is wrapped to allow for easy testing and
// easy replacement in case we run into concurrency issues.
func (m *Middleware) write(event *auditevent.AuditEvent) {
	//nolint:errcheck // TODO: We should come back to this and log the error
	m.aew.Write(event)
}

func keyFromHTTPMethodAndPath(method, path string) string {
	return fmt.Sprintf("%s:%s", method, path)
}
