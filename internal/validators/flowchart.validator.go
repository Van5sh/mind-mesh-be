package validators

import "example/hello/internal/apperrors"

func ValidateFlowchartName(name string) error {
	if err := ValidateRequiredString("flowchart name", name); err != nil {
		return apperrors.Validation("Flowchart name is required")
	}
	return ValidateMaxLength("flowchart name", name, 100)
}
