package capture

type Initiation string

const (
	NULLINITIATION Initiation = "NULL"

	// UNDEFINE Cycle was undefined
	UNDEFINE Initiation = "UNDEFINE"

	// TRIGGER A trigger occurred
	TRIGGER Initiation = "TRIGGER"

	// REPEAT_PERIOD Repeat period expired
	REPEAT_PERIOD Initiation = "REPEAT_PERIOD"

	// REQUESTED State of cycle changed to requested
	REQUESTED Initiation = "REQUESTED"
)
