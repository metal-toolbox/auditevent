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
package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeAuditEvent(t *testing.T, f *os.File, i int) {
	t.Helper()
	_, err := f.WriteString(fmt.Sprintf("audit-%d\n", i))
	require.NoError(t, err, "Unexpected error writing audit event")
}

func generateAuditEvents(t *testing.T, path string, count int) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	require.NoError(t, err, "unexpected error opening audit log file")
	defer f.Close()

	for i := 0; i < count; i++ {
		writeAuditEvent(t, f, i)
	}
}

// readAllAuditEvents reads all the events from the reader and fails if it times out
// it will read 4 bytes at a time and count newlines to determine the amount of
// audit events.
func readAllAuditEvents(t *testing.T, reader io.Reader, expectedEvents int) {
	t.Helper()
	var count int
	ticket := time.NewTicker(1 * time.Millisecond)

	for {
		select {
		case <-ticket.C:
			data := make([]byte, 4)
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

func TestTailWithoutArguments(t *testing.T) {
	t.Parallel()

	c := NewRootCmd()
	buf := bytes.NewBufferString("")
	c.SetOutput(buf)

	args := []string{}

	c.SetArgs(args)
	err := c.Execute()
	require.Error(t, err, "unexpected success")
}

func TestTailHappyPath(t *testing.T) {
	t.Parallel()

	// initialize command
	c := NewRootCmd()

	// initialize concurrent safe reader and writer
	reader, writer := io.Pipe()
	c.SetOutput(writer)

	var path string

	// initialize files
	tmpDir := t.TempDir()
	path = filepath.Join(tmpDir, "audit.log")

	// set command line arguments
	args := []string{"-f"}
	args = append(args, path)
	c.SetArgs(args)

	// Events to write
	numEvents := 100

	// generate test events
	var wg sync.WaitGroup
	wg.Add(1)
	func() {
		defer wg.Done()
		generateAuditEvents(t, path, numEvents)
	}()

	// Allow for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Run command concurrently... this will block until we cancel it
	ech := make(chan error)
	go func(ech chan error) {
		ech <- c.ExecuteContext(ctx)
	}(ech)

	// wait for all events to be written
	wg.Wait()

	// The audit file should exist
	_, serr := os.Stat(path)
	require.NoError(t, serr, "unexpected error")

	// read all the events. Fails if it times out reading them
	readAllAuditEvents(t, reader, numEvents)

	// cancel the command (this should clear the go routine)
	cancel()

	err := <-ech
	require.NoError(t, err, "unexpected error")
	require.NotEmpty(t, path, "path should not be empty")
}

func TestRootFailsCreatingNamedPipe(t *testing.T) {
	t.Parallel()

	// Initialize parent and sub-command. This is useful to inherit
	// the persistent flags.
	c := NewRootCmd()

	// Ensure buffered output in both commands
	buf := bytes.NewBufferString("")
	c.SetOutput(buf)

	// create audit file
	tmpDir := t.TempDir()
	f, terr := ioutil.TempFile(tmpDir, "audit")
	require.NoError(t, terr, "unexpected error creating temp file")

	// No execute access so mkfifo will fail
	cherr := os.Chmod(tmpDir, 0o600)
	require.NoError(t, cherr, "unexpected error changing permissions")
	// allow cleanup
	defer func() {
		cherr := os.Chmod(tmpDir, 0o700)
		require.NoError(t, cherr, "unexpected error changing permissions")
	}()

	// Set the arguments.
	args := append([]string{"-f"}, f.Name())
	perr := c.ParseFlags(args)
	require.NoError(t, perr, "unexpected error parsing flags")

	err := c.RunE(c, args)
	require.ErrorContains(t, err, "creating named pipe")
	require.Empty(t, buf.String(), "It should return an error and not write to stdout/err")
}

func TestRootFailsCreatingTailer(t *testing.T) {
	t.Parallel()

	// Initialize parent and sub-command. This is useful to inherit
	// the persistent flags.
	c := NewRootCmd()

	// Ensure buffered output in both commands
	buf := bytes.NewBufferString("")
	c.SetOutput(buf)

	// create audit file
	tmpDir := t.TempDir()
	f, terr := ioutil.TempFile(tmpDir, "audit")
	require.NoError(t, terr, "unexpected error creating temp file")

	// write and execute, can't read
	cherr := f.Chmod(0o300)
	require.NoError(t, cherr, "unexpected error changing permissions")

	// Set the arguments.
	args := append([]string{"-f"}, f.Name())

	perr := c.ParseFlags(args)
	require.NoError(t, perr, "unexpected error parsing flags")

	err := c.RunE(c, args)
	require.ErrorContains(t, err, "creating file tailer")
	require.Empty(t, buf.String(), "It should return an error and not write to stdout/err")
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("error")
}

func TestFileTrailerUnknownErrorWhenTailing(t *testing.T) {
	t.Parallel()

	er := &errorReader{}
	buf := bytes.NewBufferString("")
	ft := &fileTailer{
		r: er,
		w: buf,
	}

	// trying to tail a reader that returns an unknown error should
	// return immediately and should also surface the error
	err := ft.tailFile(context.Background())
	require.Error(t, err, "unexpected success")
}

func TestRootCmdSingletonGet(t *testing.T) {
	t.Parallel()

	c := GetCmd()
	require.Equal(t, &rootCmd, &c, "GetCmd() should return the rootCmd singleton")
}
