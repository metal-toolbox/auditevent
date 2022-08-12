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
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	ownerGroupOwnership = 0o640
)

func validateCommonArgs(cmd *cobra.Command, args []string) error {
	f, ferr := cmd.Flags().GetString("file")

	if ferr != nil {
		return ferr
	}

	if f == "" {
		return fmt.Errorf("--file is required")
	}

	return nil
}

func createNamedPipe(file string) error {
	if err := syscall.Mkfifo(file, ownerGroupOwnership); err != nil {
		// Don't fail if the file already exists.
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("creating named pipe: %w", err)
	}
	return nil
}
