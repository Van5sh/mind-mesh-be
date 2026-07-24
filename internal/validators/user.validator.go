package validators

import (
	"net/mail"
	"strings"

	"example/hello/internal/apperrors"
)

func ValidateUsername(username string) error {
	if err := ValidateRequiredString("username", username); err != nil {
		return apperrors.Validation("Incorrect UserName check the length")
	}
	return ValidateMaxLength("username", username, 50)
}

func ValidateEmail(email string) error {
	if err := ValidateRequiredString("email", email); err != nil {
		return apperrors.Validation("Email is required")
	}
	if err := ValidateMaxLength("email", email, 100); err != nil {
		return apperrors.Validation("The email is too long, maximum length is 100 characters")
	}
	_, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return apperrors.Validation("email is invalid")
	}
	return nil
}

func ValidatePasswordHash(passwordHash string) error {
	if err := ValidateRequiredString("password hash", passwordHash); err != nil {
		return apperrors.Validation("Password hash is required")
	}
	return ValidateMaxLength("password hash", passwordHash, 255)
}

func ValidateFirstName(firstName string) error {
	if err := ValidateRequiredString("first name", firstName); err != nil {
		return err
	}
	return ValidateMaxLength("first name", firstName, 50)
}

func ValidateLastName(lastName string) error {
	if err := ValidateRequiredString("last name", lastName); err != nil {
		return apperrors.Validation("Last name is required")
	}
	return ValidateMaxLength("last name", lastName, 50)
}

func ValidateUserBio(bio string) error {
	return ValidateOptionalMaxLength("bio", bio, 5000)
}

func ValidateAvatarURL(avatarURL string) error {
	return ValidateOptionalMaxLength("avatar url", avatarURL, 2048)
}
