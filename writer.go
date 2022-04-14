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
package auditevent

import (
	"encoding/json"
	"io"
)

// EventEncoder allows for encoding audit events.
// The parameter to the `Encode` method is the audit event to encode
// and it must accept pointer to an AuditEvent struct.
type EventEncoder interface {
	Encode(any) error
}

// EventWriter writes audit events to a writer using
// a given encoder.
type EventWriter struct {
	w   io.Writer
	enc EventEncoder
}

// AuditEventEncoderJSON is an encoder that encodes audit events
// using a default JSON encoder.
func NewDefaultAuditEventWriter(w io.Writer) *EventWriter {
	enc := json.NewEncoder(w)
	return NewAuditEventWriter(w, enc)
}

func NewAuditEventWriter(w io.Writer, enc EventEncoder) *EventWriter {
	return &EventWriter{w: w, enc: enc}
}

func (w *EventWriter) Write(e *AuditEvent) error {
	return w.enc.Encode(e)
}
