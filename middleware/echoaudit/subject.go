/*
Copyright 2023 Equinix, Inc.

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
package echoaudit

import (
	"github.com/labstack/echo/v4"
)

// SubjectHandler is a function that returns the AuditEvent subject map
// for a given request. This will be called after other middleware; e.g.
// the given gin context should already contain the subject information.
type SubjectHandler func(c echo.Context) map[string]string

func GetSubjectDefault(c echo.Context) map[string]string {
	// These context keys come from github.com/metal-toolbox/hollow-toolbox/ginjwt
	// TODO - these need to come from echo speciffic middlwares
	sub := getString(c, "jwt.subject")
	if sub == "" {
		sub = "Unknown"
	}

	user := getString(c, "jwt.user")
	if user == "" {
		user = c.Request().Header.Get("X-User-Id")
		if user == "" {
			user = "Unknown"
		}
	}

	return map[string]string{
		"user": user,
		"sub":  sub,
	}
}
