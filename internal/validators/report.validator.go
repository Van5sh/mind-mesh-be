package validators

import "example/hello/internal/apperrors"

func ValidateReportTitle(title string) error {
	if err := ValidateRequiredString("report title", title); err != nil {
		return apperrors.Validation("Report title is required")
	}
	return ValidateMaxLength("report title", title, 100)
}

func ValidateReportContent(content string) error {
	if err := ValidateRequiredString("report content", content); err != nil {
		return apperrors.Validation("Report content is required")
	}
	return nil
}
