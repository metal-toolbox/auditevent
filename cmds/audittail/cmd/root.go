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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	// is the default time to wait in between audit log event reads.
	defaultConstantBackoff = 5 * time.Millisecond
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = NewRootCmd()

func GetCmd() *cobra.Command {
	return rootCmd
}

func NewRootCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "audittail",
		Short: "A utility to reliably tail audit an audit log file",
		Long: `A utility to reliably tail audit an audit log file.
	
This utility will create a named pipe in order to tail the audit log file.`,
		Args: cobra.MatchAll(cobra.OnlyValidArgs, validateCommonArgs),
		RunE: tailMain,
	}

	c.PersistentFlags().StringP("file", "f", "", "audit log file to tail")
	return c
}

func tailMain(cmd *cobra.Command, args []string) error {
	//nolint:errcheck // This is already verified by cobra
	f, _ := cmd.Flags().GetString("file")

	if err := createNamedPipe(f); err != nil {
		// If the file already exists this is not an issue.
		if !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("creating named pipe: %w", err)
		}
	}

	ft, err := newFileTailer(f, cmd.OutOrStdout())
	if err != nil {
		return fmt.Errorf("creating file tailer: %w", err)
	}

	return ft.tailFile(cmd.Context())
}

type fileTailer struct {
	r io.Reader
	w io.Writer
}

func newFileTailer(file string, w io.Writer) (*fileTailer, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	return &fileTailer{
		r: f,
		w: w,
	}, nil
}

func (ft *fileTailer) tailFile(ctx context.Context) error {
	ticker := time.NewTicker(defaultConstantBackoff)
	eg := &errgroup.Group{}
	eg.Go(func() error {
		for {
			select {
			case <-ticker.C:
				if _, err := io.Copy(ft.w, ft.r); err != nil {
					if !errors.Is(err, io.EOF) {
						return err
					}
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	err := eg.Wait()
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	return fmt.Errorf("tail file: %w", err)
}
