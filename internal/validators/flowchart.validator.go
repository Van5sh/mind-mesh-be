package validators

func ValidateFlowchartName(name string) error {
	if err := ValidateRequiredString("flowchart name", name); err != nil {
		return err
	}
	return ValidateMaxLength("flowchart name", name, 100)
}
