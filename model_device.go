package capture

//Device Adapter device
type Device struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled,omitempty"`

	Properties map[string]string `json:"properties,omitempty"`
	Fields     map[string]Field  `json:"fields,omitempty"`
}
