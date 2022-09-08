package internal

import "fmt"

type PreflightTriggerCustomError struct {
	Message string
	Err     error
}

func (e *PreflightTriggerCustomError) Error() string {
	return fmt.Sprintf(e.Message + ": " + e.Err.Error())
}
