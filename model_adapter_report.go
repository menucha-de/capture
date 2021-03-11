package capture

import (
	"time"
)

type AdapterReport struct {
	ApplicationID     string          `json:"applicationId,omitempty"`
	ReportName        string          `json:"reportName,omitempty"`
	Date              time.Time       `json:"date,omitempty"`
	TotalMilliseconds int64           `json:"totalMilliseconds,omitempty"`
	Initiation        Initiation      `json:"initiation,omitempty"`
	Initiator         string          `json:"initiator,omitempty"`
	Termination       Termination     `json:"termination,omitempty"`
	Terminator        string          `json:"terminator,omitempty"`
	Devices           []*DeviceReport `json:"devices,omitempty"`
}
