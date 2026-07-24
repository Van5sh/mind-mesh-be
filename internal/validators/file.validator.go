package validators

import "example/hello/internal/apperrors"

func ValidateFolderName(name string) error {
	if err := ValidateRequiredString("folder name", name); err != nil {
		return apperrors.Validation("Folder name is required")
	}
	return ValidateMaxLength("folder name", name, 100)
}

func ValidateFileName(name string) error {
	if err := ValidateRequiredString("file name", name); err != nil {
		return apperrors.Validation("File name is required")
	}
	return ValidateMaxLength("file name", name, 255)
}

func ValidateFileSize(size int64) error {
	return ValidateNonNegativeInt64("file size", size)
}

func ValidateBucketName(bucketName string) error {
	if err := ValidateRequiredString("bucket name", bucketName); err != nil {
		return err
	}
	return ValidateMaxLength("bucket name", bucketName, 100)
}

func ValidateObjectKey(objectKey string) error {
	if err := ValidateRequiredString("object key", objectKey); err != nil {
		return err
	}
	return ValidateMaxLength("object key", objectKey, 2048)
}

func ValidateMimeType(mimeType string) error {
	if err := ValidateRequiredString("mime type", mimeType); err != nil {
		return err
	}
	return ValidateMaxLength("mime type", mimeType, 100)
}

func ValidateOriginalName(originalName string) error {
	return ValidateOptionalMaxLength("original name", originalName, 255)
}
