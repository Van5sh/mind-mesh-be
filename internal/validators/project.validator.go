package validators

import "example/hello/internal/apperrors"

func ValidateProjectName(name string) error {
	if err := ValidateRequiredString("project name", name); err != nil {
		return apperrors.Validation("Project name is required")
	}
	return ValidateMaxLength("project name", name, 100)
}

func ValidateProjectDescription(description string) error {
	if err := ValidateOptionalMaxLength("project description", description, 5000); err != nil {
		return apperrors.Validation("Project description is too long")
	}
	return nil
}
