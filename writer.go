package auditevent

import (
	"encoding/json"
	"io"
)

// EventEncoder allows for encoding audit events.
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
