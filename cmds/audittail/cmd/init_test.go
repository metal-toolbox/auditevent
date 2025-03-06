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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/metal-toolbox/auditevent/internal/testtools"
)

func TestInit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		args               []string
		appendAuditLogFile bool
		wantErr            bool
	}{
		{
			name: "should initialize pipe",
			args: []string{
				"init", "-f",
			},
			appendAuditLogFile: true,
			wantErr:            false,
		},
		{
			name: "no arguments should fail",
			args: []string{
				"init",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewRootCmd()
			buf := bytes.NewBufferString("")
			c.SetOutput(buf)
			c.AddCommand(NewInitCommand())

			args := tt.args
			var path string

			if tt.appendAuditLogFile {
				tmpDir := t.TempDir()
				path = filepath.Join(tmpDir, "audit.log")
				args = append(args, path)
			}

			c.SetArgs(args)
			err := c.Execute()
			if tt.wantErr {
				require.Error(t, err, "unexpected success")
				return
			}

			require.NotEmpty(t, path, "path should not be empty")
			require.NoError(t, err, "unexpected error")

			_, serr := os.Stat(path)
			require.NoError(t, serr, "unexpected error")
			require.Contains(t, buf.String(), "Created named pipe")
		})
	}
}

func TestInitTailFileFailsIfItCantCreateFIFO(t *testing.T) {
	t.Parallel()

	// Initialize parent and sub-command. This is useful to inherit
	// the persistent flags.
	c := NewRootCmd()
	initCmd := NewInitCommand()
	c.AddCommand(initCmd)

	// Ensure buffered output in both commands
	buf := bytes.NewBufferString("")
	c.SetOutput(buf)
	initCmd.SetOutput(buf)

	// Set the arguments
	args := append([]string{"-f"}, "/foo/bar/")

	perr := initCmd.ParseFlags(args)
	require.NoError(t, perr)

	err := initCmd.RunE(initCmd, args)
	require.Error(t, err, "unexpected success")
	require.Empty(t, buf.String(), "It should return an error and not write to stdout/err")
}

func TestInitTailFileFailsToWriteSuccess(t *testing.T) {
	t.Parallel()

	// Initialize parent and sub-command. This is useful to inherit
	// the persistent flags.
	c := NewRootCmd()
	initCmd := NewInitCommand()
	c.AddCommand(initCmd)

	// Create error from output
	ew := testtools.NewErrorWriter()
	c.SetOutput(ew)
	initCmd.SetOutput(ew)

	// Set the arguments
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "audit.log")
	args := append([]string{"-f"}, path)

	perr := initCmd.ParseFlags(args)
	require.NoError(t, perr)

	err := initCmd.RunE(initCmd, args)
	require.ErrorContains(t, err, "writing to stdout",
		"there should be an error and it should contain the expected message")
}

func TestInitSucceedsEvenIfFileAlreadyExists(t *testing.T) {
	t.Parallel()

	c := NewRootCmd()
	buf := bytes.NewBufferString("")
	c.SetOutput(buf)
	c.AddCommand(NewInitCommand())

	args := []string{"init", "-f"}
	var path string

	tmpDir := t.TempDir()
	path = filepath.Join(tmpDir, "audit.log")
	args = append(args, path)

	c.SetArgs(args)
	err := c.Execute()
	require.NoError(t, err, "unexpected error")

	_, serr := os.Stat(path)
	require.NoError(t, serr, "unexpected error")
	require.Contains(t, buf.String(), "Created named pipe")

	// A second call should still succeed
	c.SetArgs(args)
	err = c.Execute()
	require.NoError(t, err, "unexpected error in second call")
}
