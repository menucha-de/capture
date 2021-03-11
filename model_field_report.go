package capture

import "time"

type FieldReport struct {
	Date  time.Time   `json:"date,omitempty"`
	Name  string      `json:"name,omitempty"`
	Value interface{} `json:"value,omitempty"`
}
