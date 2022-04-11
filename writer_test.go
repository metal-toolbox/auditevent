package auditevent_test

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent"
)

func getNamedPipe(t *testing.T) string {
	t.Helper()

	dirName, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	pipeName := dirName + "/test.pipe"

	err = syscall.Mkfifo(pipeName, 0o600)
	require.NoError(t, err)

	return pipeName
}

func TestEventIsSuccessfullyWritten(t *testing.T) {
	t.Parallel()

	expectedEvent := auditevent.NewAuditEvent(
		"UserCreate",
		auditevent.EventSource{
			Type:  "IP",
			Value: "127.0.0.1",
		},
		"Success",
		map[string]string{
			"username": "test",
		},
		"test-component",
	).WithTarget(map[string]string{
		"path":    "/user",
		"newUser": "foobar",
	})

	nppath := getNamedPipe(t)

	// writer
	go func(eventToWrite *auditevent.AuditEvent) {
		np, err := os.OpenFile(nppath, os.O_WRONLY, 0o600)
		require.NoError(t, err)
		defer np.Close()

		aew := auditevent.NewDefaultAuditEventWriter(np)
		err = aew.Write(eventToWrite)
		require.NoError(t, err)
	}(expectedEvent)

	// reader
	fd, err := os.Open(nppath)
	require.NoError(t, err)

	// This will block until the writer goroutine writes the event
	rawEvent, err := io.ReadAll(fd)
	require.NoError(t, err)

	var gotEvent auditevent.AuditEvent
	umerr := json.Unmarshal(rawEvent, &gotEvent)
	require.NoError(t, umerr)

	require.Equal(t, expectedEvent.Metadata.AuditID, gotEvent.Metadata.AuditID, "audit metadata should match")
	require.Equal(t, expectedEvent.Type, gotEvent.Type, "type should match")
	require.True(t, expectedEvent.LoggedAt.Equal(gotEvent.LoggedAt), "logging time should match")
	require.True(t, expectedEvent.LoggedAt.Before(time.Now()), "logging time should be before now")
	require.Equal(t, expectedEvent.Source.Type, gotEvent.Source.Type, "source type should match")
	require.Equal(t, expectedEvent.Source.Value, gotEvent.Source.Value, "source value should match")
	require.Equal(t, expectedEvent.Outcome, gotEvent.Outcome, "outcome should match")
	require.Equal(t, expectedEvent.Subjects, gotEvent.Subjects, "subjects should match")
	require.Equal(t, expectedEvent.Component, gotEvent.Component, "component should match")
	require.Equal(t, expectedEvent.Target, gotEvent.Target, "target should match")
}
