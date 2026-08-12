package validators

import (
	"encoding/json"

	"example/hello/internal/apperrors"
)

func ValidateFlowchartName(name string) error {
	if err := ValidateRequiredString("flowchart name", name); err != nil {
		return apperrors.Validation("Flowchart name is required")
	}

	return ValidateMaxLength("flowchart name", name, 100)
}

func ValidateFlowchartData(data []byte) error {
	if len(data) == 0 {
		return apperrors.Validation("Flowchart data is required")
	}

	if !json.Valid(data) {
		return apperrors.Validation("Invalid flowchart JSON")
	}

	return nil
}
