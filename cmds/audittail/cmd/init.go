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
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = NewInitCommand()

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize an audit log file to tail",
		Long: `Initialize an audit log file to tail.

This initializes an audit log file as a named pipe in order to tail it.
This is useful to run in an init container to ensure the file is ready.`,
		Args: cobra.MatchAll(cobra.OnlyValidArgs, validateCommonArgs),
		RunE: initTailFile,
	}
}

//nolint:gochecknoinits // this is a practice recommended by cobra
func init() {
	rootCmd.AddCommand(initCmd)
}

func initTailFile(cmd *cobra.Command, _ []string) error {
	//nolint:errcheck // This is already verified by cobra
	f, _ := cmd.Flags().GetString("file")

	if err := createNamedPipe(f); err != nil {
		return err
	}

	_, werr := fmt.Fprintf(cmd.OutOrStdout(), "Created named pipe %s\n", f)
	if werr != nil {
		return fmt.Errorf("writing to stdout: %w", werr)
	}

	return nil
}
