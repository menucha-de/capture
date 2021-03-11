package capture

type State string

const (
	//UNDEFINED Cycle is undefined
	UNDEFINED State = "UNDEFINED"

	//UNREQUESTED Cycle is unrequested
	UNREQUESTED State = "UNREQUESTED"

	//STATEREQUESTED Cycle is requested
	STATEREQUESTED State = "REQUESTED"

	// ACTIVE Cycle is active
	ACTIVE State = "ACTIVE"
)
