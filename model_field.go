package capture

//Field Device Field
type Field struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Label      string            `json:"label"`
	Properties map[string]string `json:"properties,omitempty"`
	Value      interface{}       `json:"value,omitempty"`
}
