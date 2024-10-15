package repeatable

import (
	"fmt"
	"strings"
)

func ValidateRequiredFields(fields map[string]string) error {
	for fieldName, value := range fields {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("field %s is required", fieldName)
		}
	}
	return nil
}
