package validators

func ValidateProjectName(name string) error {
	if err := ValidateRequiredString("project name", name); err != nil {
		return err
	}
	return ValidateMaxLength("project name", name, 100)
}

func ValidateProjectDescription(description string) error {
	return ValidateOptionalMaxLength("project description", description, 5000)
}
