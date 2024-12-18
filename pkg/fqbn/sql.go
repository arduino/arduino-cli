package fqbn

import "fmt"

// Value implements the driver.Valuer interface for the FQBN type.
func (f FQBN) Value() (any, error) {
	return f.String(), nil
}

// Scan implements the sql.Scanner interface for the FQBN type.
func (f *FQBN) Scan(value any) error {
	if value == nil {
		return nil
	}

	if v, ok := value.(string); ok {
		ParsedFQBN, err := Parse(v)
		if err != nil {
			return fmt.Errorf("failed to parse FQBN: %w", err)
		}
		*f = *ParsedFQBN
		return nil
	}

	return fmt.Errorf("unsupported type: %T", value)
}
