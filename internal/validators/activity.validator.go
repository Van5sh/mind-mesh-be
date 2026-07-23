package validators

func ValidateActivityAction(action string) error {
	if err := ValidateRequiredString("activity action", action); err != nil {
		return err
	}
	return ValidateMaxLength("activity action", action, 100)
}

func ValidateEntityType(entityType string) error {
	return ValidateOptionalMaxLength("entity type", entityType, 50)
}
