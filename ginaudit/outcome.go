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
package ginaudit

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/metal-toolbox/auditevent"
)

// OutcomeHandler is a function that returns the AuditEvent outcome
// for a given request. This will be called after other middleware; e.g.
// the given gin context should already contain a result status.
// It is recommended to return one of the samples defined in
// `samples.go`.
type OutcomeHandler func(c *gin.Context) string

// GetOutcomeDefault is the default outcome handler that's set in
// the middleware constructor. It will return `failed` for HTTP response
// statuses 500 and above, `denied` for requests 400 and above and
// `succeeded` otherwise.
func GetOutcomeDefault(c *gin.Context) string {
	status := c.Writer.Status()
	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		return auditevent.OutcomeDenied
	}
	if status >= http.StatusInternalServerError {
		return auditevent.OutcomeFailed
	}
	return auditevent.OutcomeSucceeded
}
