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
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an audit event.
type AuditEvent struct {
	Metadata EventMetadata `json:"metadata"`
	// Type: Defines the type of event that occurred
	// This is a small identifier to quickly determine what happened.
	// e.g. UserLogin, UserLogout, UserCreate, UserDelete, etc.
	Type string `json:"type"`
	// LoggedAt: determines when the event occurred.
	// Note that this should have sufficient information to authoritatively
	// determine the exact time the event was logged at. The output must be in
	// Coordinated Universal Time (UTC) format, a modern continuation of
	// Greenwich Mean Time (GMT), or local time with an offset from UTC to satisfy
	// NIST SP 800-53 requirement AU-8.
	LoggedAt time.Time `json:"loggedAt"`
	// Source: determines the source of the event.
	// Normally, using the IP address of the client, or pod name is sufficient.
	// One must be careful of the data that's added here as we don't want to
	// leak Personally Identifiable Information.
	Source EventSource `json:"source"`
	// Outcome: determines whether the event was successful or not, e.g. successful login
	// It may also determine if the event was approved or denied.
	Outcome string `json:"outcome"`
	// Subject: is the identity of the subject of the event.
	// e.g. who triggered the event? Additional information
	// may be added, such as group membership and/or role
	Subjects map[string]string `json:"subjects"`
	// Component: allows to determine in which component the event occurred
	// (Answering the "Where" question of section c in the NIST SP 800-53
	// Revision 5.1 Control AU-3).
	Component string `json:"component"`
	// Target: Defines where the target of the operation. e.g. the path of
	// the REST resource
	// (Answering the "Where" question of section c in the NIST SP 800-53
	// Revision 5.1 Control AU-3 as well as indicating an entity
	// associated for section f).
	Target map[string]string `json:"target,omitempty"`
	// Data: enhances the audit event with extra information that may be
	// useful for forensic analysis.
	Data *json.RawMessage `json:"data,omitempty"`
}

type EventMetadata struct {
	// AuditID: is a unique identifier for the audit event.
	AuditID string `json:"auditId"`
	// Extra allows for including additional information about the event
	// that aids in tracking, parsing or auditing
	Extra map[string]any `json:"extra,omitempty"`
}

type EventSource struct {
	// Type indicates the source type. e.g. Network, File, local, etc.
	// The intent is to determine where a request came from.
	Type string `json:"type"`
	// Value aims to indicate the source of the event. e.g. IP address,
	// hostname, etc.
	Value string `json:"value"`
	// Extra allows for including additional information about the event
	// source that aids in tracking, parsing or auditing
	Extra map[string]any `json:"extra,omitempty"`
}

// NewAuditEvent returns a new AuditEvent with an appropriately set AuditID and logging time.
func NewAuditEvent(
	eventType string,
	source EventSource,
	outcome string,
	subjects map[string]string,
	component string,
) *AuditEvent {
	return &AuditEvent{
		Metadata: EventMetadata{
			AuditID: uuid.New().String(),
		},
		Type:      eventType,
		LoggedAt:  time.Now().UTC(),
		Source:    source,
		Outcome:   outcome,
		Subjects:  subjects,
		Component: component,
	}
}

// NewAuditEventWithID returns a new AuditEvent with the passed AuditID.
func NewAuditEventWithID(
	auditID string,
	eventType string,
	source EventSource,
	outcome string,
	subjects map[string]string,
	component string,
) *AuditEvent {
	return &AuditEvent{
		Metadata: EventMetadata{
			AuditID: auditID,
		},
		Type:      eventType,
		LoggedAt:  time.Now().UTC(),
		Source:    source,
		Outcome:   outcome,
		Subjects:  subjects,
		Component: component,
	}
}

// WithTarget sets the target of the event.
func (e *AuditEvent) WithTarget(target map[string]string) *AuditEvent {
	e.Target = target
	return e
}

// WithData sets the data of the event.
func (e *AuditEvent) WithData(data *json.RawMessage) *AuditEvent {
	e.Data = data
	return e
}

// WithDataFromString sets the data of the event from a string.
// Note that validating that this is properly JSON-formatted
// is the responsibility of the caller.
func (e *AuditEvent) WithDataFromString(data string) *AuditEvent {
	rawMsg := json.RawMessage(data)
	return e.WithData(&rawMsg)
}
