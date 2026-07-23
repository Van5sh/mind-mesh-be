package validators

func ValidateChatTitle(title string) error {
	return ValidateOptionalMaxLength("chat title", title, 100)
}

func ValidateChatMessageContent(content string) error {
	return ValidateRequiredString("message content", content)
}
