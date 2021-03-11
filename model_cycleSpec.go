package capture

type CycleSpec struct {
	ID                     string              `json:"id,omitempty"`
	ApplicationID          string              `json:"applicationId,omitempty"`
	Name                   string              `json:"name,omitempty"`
	Enabled                bool                `json:"enabled,omitempty"`
	Duration               int64               `json:"duration"`
	RepeatPeriod           int64               `json:"repeatPeriod"`
	Interval               int                 `json:"interval,omitempty"`
	WhenDataAvailable      bool                `json:"whenDataAvailable,omitempty"`
	WhenDataAvailableDelay int                 `json:"WhenDataAvailableDelay,omitempty"`
	ReportIfEmpty          bool                `json:"reportIfEmpty,omitempty"`
	FieldSubscriptions     map[string][]string `json:"fieldSubscriptions"`
}
