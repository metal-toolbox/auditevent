package ginaudit

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/metal-toolbox/auditevent"
)

type Middleware struct {
	component    string
	aew          *auditevent.EventWriter
	eventTypeMap sync.Map
}

const (
	ownerGroupAccess = 0o640
	retryInterval    = 100 * time.Millisecond
)

func OpenAuditLogFileUntilSuccess(path string) (*os.File, error) {
	for {
		// This is opened with the O_APPEND option to ensure
		// atomicity of writes. This is important to ensure
		// we can concurrently write to the file and not block
		// the server's main loop.
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, ownerGroupAccess)
		if err != nil {
			if os.IsNotExist(err) {
				time.Sleep(retryInterval)
				continue
			}
			// Not being able to write audit log events is a fatal error
			return nil, err
		}
		return fd, nil
	}
}

// NewMiddleware returns a new middleware instance with a default writer
// It assumes that audit events are less than 4096 bytes to ensure atomicity.
// it takes a writer for the audit log.
func NewMiddleware(component string, w io.Writer) *Middleware {
	return &Middleware{
		component: component,
		aew:       auditevent.NewDefaultAuditEventWriter(w),
	}
}

func (m *Middleware) RegisterEventType(eventType, httpMethod, path string) {
	m.eventTypeMap.Store(keyFromHTTPMethodAndPath(httpMethod, path), eventType)
}

func (m *Middleware) Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// We audit after the request has been processed
		c.Next()

		event := auditevent.NewAuditEvent(
			m.getEventType(method, path),
			auditevent.EventSource{
				Type: "IP",
				// This already takes into account X-Forwarded-For and alike headers
				Value: c.ClientIP(),
			},
			m.getOutcome(c),
			m.getSubject(c),
			m.component,
		).WithTarget(map[string]string{
			"path": path,
		})

		// persist event
		m.write(event)
	}
}

func (m *Middleware) getEventType(httpMethod, path string) string {
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

func (m *Middleware) getOutcome(c *gin.Context) string {
	status := c.Writer.Status()
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		return auditevent.OutcomeDenied
	}
	if status >= http.StatusInternalServerError {
		return auditevent.OutcomeFailed
	}
	return auditevent.OutcomeSucceeded
}

func (m *Middleware) getSubject(c *gin.Context) map[string]string {
	// These context keys come from github.com/metal-toolbox/hollow-toolbox/ginjwt
	sub := c.GetString("jwt.subject")
	if sub == "" {
		sub = "Unknown"
	}

	user := c.GetString("jwt.user")
	if user == "" {
		user = c.Request.Header.Get("X-User-Id")
		if user == "" {
			user = "Unknown"
		}
	}
	return map[string]string{
		"user": user,
		"sub":  sub,
	}
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
