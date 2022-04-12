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
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func Test_validateCommonArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "should succeed",
			args: []string{
				"-f", "foo",
			},
			wantErr: false,
		},
		{
			name: "should fail with empty file path",
			args: []string{
				"-f", "",
			},
			wantErr: true,
		},
		{
			name:    "should fail without arguments",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := NewRootCmd()
			c.SetArgs(tt.args)
			perr := c.ParseFlags(tt.args)
			require.NoError(t, perr)
			err := validateCommonArgs(c, tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func Test_validateCommonArgsFailsWithoutArguments(t *testing.T) {
	t.Parallel()

	c := &cobra.Command{}
	err := validateCommonArgs(c, []string{})
	require.Error(t, err)
}
