package capture

type Termination string

const (
	NULLTERMINATION Termination = "NULL"

	//UNDEFINETERMINATION Cycle was undefined
	UNDEFINETERMINATION Termination = "UNDEFINE"

	//TRIGGERTERMINATION A trigger occurred
	TRIGGERTERMINATION Termination = "TRIGGER"

	//DURATION Duration period expired
	DURATION Termination = "DURATION"

	//INTERVAL Interval time expired
	INTERVAL Termination = "INTERVAL"

	//DATAAVAILABLE New data available
	DATAAVAILABLE Termination = "DATA_AVAILABLE"

	//UNREQUESTEDTERMINATION State of cycle changed to unrequested
	UNREQUESTEDTERMINATION Termination = "UNREQUESTED"

	//ERROR Error occurred
	ERROR Termination = "ERROR"
)
