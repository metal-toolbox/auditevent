package ginaudit_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent"
	"github.com/metal-toolbox/auditevent/ginaudit"
)

const (
	comp = "test"
)

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
				"POST:/fails-with-user-header",
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
	}
}

func getNamedPipe(t *testing.T) string {
	t.Helper()

	dirName := t.TempDir()

	pipeName := dirName + "/test.pipe"

	err := syscall.Mkfifo(pipeName, 0o600)
	require.NoError(t, err)

	return pipeName
}

func setFixtures(t *testing.T, w io.Writer) *gin.Engine {
	t.Helper()

	mdw := ginaudit.NewJSONMiddleware(comp, w)

	r := gin.New()
	r.Use(mdw.Audit())

	// allowed user with registered event type
	mdw.RegisterEventType("MyEventType", "GET", "/ok")
	r.GET("/ok", func(c *gin.Context) {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		c.JSON(http.StatusOK, "ok")
	})

	// Writing to `/ok` breaks the app
	r.POST("/fails-with-user-header", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	})

	// denied with no user
	r.GET("/denied", func(c *gin.Context) {
		c.JSON(http.StatusForbidden, "denied")
	})

	// denied with user
	r.GET("/denied-user", func(c *gin.Context) {
		c.Set("jwt.user", "user-ozz")
		c.Set("jwt.subject", "sub-ozz")
		c.JSON(http.StatusForbidden, "denied")
	})

	return r
}

func setPipeReader(t *testing.T, namedPipe string) <-chan io.WriteCloser {
	t.Helper()
	rchan := make(chan io.WriteCloser)
	go func(c chan<- io.WriteCloser) {
		fd, err := ginaudit.OpenAuditLogFileUntilSuccess(namedPipe)
		require.NoError(t, err)
		c <- fd
	}(rchan)
	return rchan
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	for _, tc := range getTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := getNamedPipe(t)

			fdchan := setPipeReader(t, p)

			f, err := os.Open(p)
			require.NoError(t, err)

			// receive pipe reader file descriptor
			pfd := <-fdchan
			defer pfd.Close()

			r := setFixtures(t, pfd)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.expectedEvent.Target["path"], nil)
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
		})
	}
}

func TestParallelCallsToMiddleware(t *testing.T) {
	t.Parallel()

	// Set up server with middleware
	p := getNamedPipe(t)
	c := setPipeReader(t, p)

	// set up other end of pipe
	f, err := os.OpenFile(p, os.O_RDONLY|os.O_APPEND, 0o600)
	require.NoError(t, err)
	defer f.Close()

	// receive pipe reader file descriptor
	pfd := <-c

	r := setFixtures(t, pfd)

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
			req := httptest.NewRequest(method, path, nil)
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
}

func TestOpenAuditLogFileUntilSuccess(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)

	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "audit.log")

	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		fd, err := os.OpenFile(tmpfile, os.O_RDONLY|os.O_CREATE, 0o600)
		require.NoError(t, err)
		err = fd.Close()
		require.NoError(t, err)
	}()

	fd, err := ginaudit.OpenAuditLogFileUntilSuccess(tmpfile)
	require.NoError(t, err)
	require.NotNil(t, fd)

	err = fd.Close()
	require.NoError(t, err)

	// We wait so we don't leak file descriptors
	wg.Wait()

	err = os.Remove(tmpfile)
	require.NoError(t, err)
}

func TestOpenAuditLogFileError(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	wg.Add(1)

	tmpdir := t.TempDir()
	tmpfile := filepath.Join(tmpdir, "audit.log")

	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		// This file is read only
		fd, err := os.OpenFile(tmpfile, os.O_RDONLY|os.O_CREATE, 0o500)
		require.NoError(t, err)
		err = fd.Close()
		require.NoError(t, err)
	}()

	fd, err := ginaudit.OpenAuditLogFileUntilSuccess(tmpfile)
	require.Error(t, err)
	require.Nil(t, fd)

	// We wait so we don't leak file descriptors
	wg.Wait()

	err = os.Remove(tmpfile)
	require.NoError(t, err)
}
