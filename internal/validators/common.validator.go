package validators

import (
	"strings"
	"unicode/utf8"

	"example/hello/internal/apperrors"

	"github.com/jackc/pgx/v5/pgtype"
)

func ValidateRequiredString(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.Validation(fieldName + " is required")
	}
	return nil
}

func ValidateMaxLength(fieldName, value string, max int) error {
	if utf8.RuneCountInString(value) > max {
		return apperrors.Validation(fieldName + " must be at most " + itoa(max) + " characters")
	}
	return nil
}

func ValidateMinLength(fieldName, value string, min int) error {
	if utf8.RuneCountInString(value) < min {
		return apperrors.Validation(fieldName + " must be at least " + itoa(min) + " characters")
	}
	return nil
}

func ValidateOptionalMaxLength(fieldName, value string, max int) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return ValidateMaxLength(fieldName, value, max)
}

func ValidateNonNegativeInt64(fieldName string, value int64) error {
	if value < 0 {
		return apperrors.Validation(fieldName + " must be non-negative")
	}
	return nil
}

func ValidateUUID(fieldName string, value pgtype.UUID) error {
	if !value.Valid {
		return apperrors.Validation(fieldName + " is required")
	}
	return nil
}

func ValidateUUIDSlice(fieldName string, values []pgtype.UUID, allowEmpty bool) error {
	if !allowEmpty && len(values) == 0 {
		return apperrors.Validation(fieldName + " must not be empty")
	}
	for _, value := range values {
		if !value.Valid {
			return apperrors.Validation(fieldName + " contains an invalid id")
		}
	}
	return nil
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [20]byte
	index := len(digits)
	for value > 0 {
		index--
		digits[index] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[index:])
}
