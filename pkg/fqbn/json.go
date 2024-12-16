package fqbn

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON implements the json.Unmarshaler interface for the FQBN type.
func (f *FQBN) UnmarshalJSON(data []byte) error {
	var fqbnStr string
	if err := json.Unmarshal(data, &fqbnStr); err != nil {
		return fmt.Errorf("failed to unmarshal FQBN: %w", err)
	}

	fqbn, err := Parse(fqbnStr)
	if err != nil {
		return fmt.Errorf("invalid FQBN: %w", err)
	}

	*f = *fqbn
	return nil
}

// MarshalJSON implements the json.Marshaler interface for the FQBN type.
func (f FQBN) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}
