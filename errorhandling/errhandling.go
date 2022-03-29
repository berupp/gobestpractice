package errorhandling

import "fmt"

var (
	//ConnectionError is a package defined error that allows users to react to different error conditions
	ConnectionError = fmt.Errorf("connection failed")
)

type CustomError struct {
	Status int
	Reason string
}

//Error satisfies the error interface. Be aware what you return if you implement with or without pointer receiver
func (e CustomError) Error() string {
	return fmt.Sprintf("failed with status %d: %s", e.Status, e.Reason)
}

func ReturnCustomError() error {
	//Returning CustomError value (not pointer) as Error() is implemented with value receiver
	return CustomError{
		Status: 22,
		Reason: "Just cause",
	}
}

func ReturnPredefinedError() error {
	return ConnectionError
}
