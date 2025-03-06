//go:build testtools
// +build testtools

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
package testtools

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent/helpers"
)

const (
	ownerAccessOnly   = 0o600
	ownerWriteAllRead = 0o644
	oneWord           = 4
)

// GetNamedPipe creates a randomly named pipe in a temporary directory.
// This function returns the path to the pipe. The pipe's lifetime is
// akin to the test's lifetime and will be cleaned up upon the test ending.
func GetNamedPipe(t *testing.T) string {
	t.Helper()

	dirName := t.TempDir()

	pipeName := dirName + "/test.pipe"

	err := syscall.Mkfifo(pipeName, ownerAccessOnly)
	require.NoError(t, err)

	return pipeName
}

// SetPipeReader creates creates a reader that will return a file
// descriptor to a named pipe whenever possible. This function exists given
// that named pipes are blocking, and so, it allows for spawning the other
// end (the writer) of the pipe in parallel.
func SetPipeReader(t *testing.T, namedPipe string) <-chan io.WriteCloser {
	t.Helper()
	rchan := make(chan io.WriteCloser)
	go func(c chan<- io.WriteCloser) {
		fd, err := helpers.OpenAuditLogFileUntilSuccess(namedPipe)
		require.NoError(t, err)
		c <- fd
	}(rchan)
	return rchan
}

// WriteAuditEvent writes a test audit event to a file.
func WriteAuditEvent(t *testing.T, f *os.File, i int) {
	t.Helper()
	_, err := fmt.Fprintf(f, "audit-%d\n", i)
	require.NoError(t, err, "Unexpected error writing audit event")
}

// GenerateAuditEvents writes a number of audit events to a file.
func GenerateAuditEvents(t *testing.T, path string, count int) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, ownerWriteAllRead)
	require.NoError(t, err, "unexpected error opening audit log file")
	defer f.Close()

	for i := 0; i < count; i++ {
		WriteAuditEvent(t, f, i)
	}
}

// ReadAllAuditEvents reads all the events from the reader and fails if it times out
// it will read 4 bytes at a time and count newlines to determine the amount of
// audit events.
func ReadAllAuditEvents(t *testing.T, reader io.Reader, expectedEvents int) {
	t.Helper()
	var count int
	ticket := time.NewTicker(1 * time.Millisecond)

	for {
		select {
		case <-ticket.C:
			data := make([]byte, oneWord)
			_, err := reader.Read(data)
			// ignore EOF as the tail writer might not be ready with events
			if len(data) == 0 || errors.Is(err, io.EOF) {
				break
			}
			str := string(data)
			count += strings.Count(str, "\n")
			if count == expectedEvents {
				return
			}
		case <-time.After(1 * time.Second):
			require.Fail(t, "timeout. We didn't receive all the events")
		}
	}
}

// ErrorWriter is a writer that always returns an error.
type ErrorWriter struct{}

func NewErrorWriter() io.Writer {
	return &ErrorWriter{}
}

func (e *ErrorWriter) Write(_ []byte) (n int, err error) {
	return 0, fmt.Errorf("error") //nolint:err113 //test
}
