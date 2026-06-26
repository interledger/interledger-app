package configa

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var vld = validator.New()

// validateStruct runs go-playground/validator on v using its validate struct tags.
// Non-struct values are silently skipped (validator.InvalidValidationError).
func validateStruct(v any) error {
	if err := vld.Struct(v); err != nil {
		var invalidErr *validator.InvalidValidationError
		if errors.As(err, &invalidErr) {
			return nil
		}
		return fmt.Errorf("configa: %w", err)
	}
	return nil
}
