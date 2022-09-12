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

/*
The `AuditEvent` structure is used to represent an audit event.
It provides the minimal information needed to audit an event, as well as
a uniform format to persist the events in audit logs.

It is highly recommended to use the `NewAuditEvent` function to create
audit events and set the required fields.
*/
package auditevent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAuditEventWithID(t *testing.T) {
	t.Parallel()

	type args struct {
		auditID   string
		eventType string
		source    EventSource
		outcome   string
		subjects  map[string]string
		component string
	}

	tt := args{auditID: "7c96380f-24e6-4fdb-8612-d50c3f1a9806"}
	wantID := "7c96380f-24e6-4fdb-8612-d50c3f1a9806"
	got := NewAuditEventWithID(
		tt.auditID,
		tt.eventType,
		tt.source,
		tt.outcome,
		tt.subjects,
		tt.component,
	)
	require.NotNil(t, got.Metadata)
	require.Equal(t, got.Metadata.AuditID, wantID)
}
