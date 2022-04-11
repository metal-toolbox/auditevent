package auditevent

import "encoding/json"

func (e *AuditEvent) MarshalJSON() ([]byte, error) {
	type Alias AuditEvent
	return json.Marshal(&struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  e.Type,
		Alias: (*Alias)(e),
	})
}
