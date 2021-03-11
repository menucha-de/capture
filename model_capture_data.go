package capture

import (
	"time"
)

type CaptureData struct {
	Date   time.Time   `json:"date,omitempty"`
	Device string      `json:"device,omitempty"`
	Field  string      `json:"field,omitempty"`
	Value  interface{} `json:"Value,omitempty"`
}
