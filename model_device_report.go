package capture

type DeviceReport struct {
	Name   string         `json:"name,omitempty"`
	Fields []*FieldReport `json:"fields,omitempty"`
}
