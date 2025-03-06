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
package helpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

const (
	ownerGroupAccess = 0o640
	retryInterval    = 100 * time.Millisecond
)

// ErrUnexpectedAuditLog is returned when an unexpected error occurs while opening the audit log file.
var ErrUnexpectedAuditLog = errors.New("unexpected audit log error")

// OpenAuditLogFileUntilSuccess attempts to open a file for writing audit events until
// it succeeds.
// It assumes that audit events are less than 4096 bytes to ensure atomicity.
// it takes a writer for the audit log.
func OpenAuditLogFileUntilSuccess(path string, loggers ...logr.Logger) (*os.File, error) {
	l, err := newLogger(loggers...)
	if err != nil {
		return nil, err
	}

	l.Info("opening audit log file. This will block until the file is available", "path", path)

	for {
		// This is opened with the O_APPEND option to ensure
		// atomicity of writes. This is important to ensure
		// we can concurrently write to the file and not block
		// the server's main loop.
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, ownerGroupAccess)
		if err != nil {
			if os.IsNotExist(err) {
				time.Sleep(retryInterval)
				continue
			}
			// Not being able to write audit log events is a fatal error
			return nil, err
		}

		l.Info("audit log file opened successfully", "path", path)
		return fd, nil
	}
}

// OpenAuditLogFileUntilSuccessWithContext attempts to open a file for writing audit events until
// it succeeds or the context is cancelled.
// It assumes that audit events are less than 4096 bytes to ensure atomicity.
// it takes a writer for the audit log.
func OpenAuditLogFileUntilSuccessWithContext(ctx context.Context, path string, ls ...logr.Logger) (*os.File, error) {
	l, err := newLogger(ls...)
	if err != nil {
		return nil, err
	}

	l.Info("opening audit log file. This will block until the file is available or context is cancelled", "path", path)

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// This is opened with the O_APPEND option to ensure
		// atomicity of writes. This is important to ensure
		// we can concurrently write to the file and not block
		// the server's main loop.
		fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, ownerGroupAccess)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			// Not being able to write audit log events is a fatal error
			return nil, err
		}

		l.Info("audit log file opened successfully", "path", path)
		return fd, nil
	}

	return nil, ErrUnexpectedAuditLog
}

// OpenOrCreateAuditLogFile attempts to open a file for writing audit events.
// If the file does not exist, it will be created.
func OpenOrCreateAuditLogFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, ownerGroupAccess)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}
	return f, nil
}

func newLogger(loggers ...logr.Logger) (logr.Logger, error) {
	if len(loggers) > 0 {
		return loggers[0], nil
	}

	z, err := zap.NewProduction()
	if err != nil {
		return logr.Logger{}, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return zapr.NewLogger(z), nil
}
