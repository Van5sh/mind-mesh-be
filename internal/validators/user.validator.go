package validators

import (
	"net/mail"
	"strings"

	"example/hello/internal/apperrors"
)

func ValidateUsername(username string) error {
	if err := ValidateRequiredString("username", username); err != nil {
		return err
	}
	return ValidateMaxLength("username", username, 50)
}

func ValidateEmail(email string) error {
	if err := ValidateRequiredString("email", email); err != nil {
		return err
	}
	if err := ValidateMaxLength("email", email, 100); err != nil {
		return err
	}
	_, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return apperrors.Validation("email is invalid")
	}
	return nil
}

func ValidatePasswordHash(passwordHash string) error {
	if err := ValidateRequiredString("password hash", passwordHash); err != nil {
		return err
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
		return err
	}
	return ValidateMaxLength("last name", lastName, 50)
}

func ValidateUserBio(bio string) error {
	return ValidateOptionalMaxLength("bio", bio, 5000)
}

func ValidateAvatarURL(avatarURL string) error {
	return ValidateOptionalMaxLength("avatar url", avatarURL, 2048)
}
