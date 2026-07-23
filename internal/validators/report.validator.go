package validators

func ValidateReportTitle(title string) error {
	if err := ValidateRequiredString("report title", title); err != nil {
		return err
	}
	return ValidateMaxLength("report title", title, 100)
}

func ValidateReportContent(content string) error {
	return ValidateRequiredString("report content", content)
}
