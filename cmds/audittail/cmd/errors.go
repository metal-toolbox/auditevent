package cmd

import "errors"

// ErrFileRequired is returned when the --file flag is not provided.
var ErrFileRequired = errors.New("--file is required")
