package common

import "fmt"
import "encoding/json"

// ErrorMessage represents an Error with an attached message.
//
// It's the same as a normal error, but It is also parsable as JSON.
type ErrorMessage struct {
	message string
}

// MarshalJSON allows to marshal this object as a JSON object.
func (err ErrorMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.message)
}

// Error returns the error message.
func (err ErrorMessage) Error() string {
	return fmt.Sprint(err.message)
}

// String returns a string representation of the Error.
func (err ErrorMessage) String() string {
	return err.Error()
}

// FromError creates an ErrorMessage from an Error.
func FromError(err error) ErrorMessage {
	return ErrorMessage{
		message: err.Error(),
	}
}
