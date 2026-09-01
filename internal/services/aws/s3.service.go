package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service(client *s3.Client, bucket string) *S3Service {
	return &S3Service{
		client: client,
		bucket: bucket,
	}
}

func (s *S3Service) S3Upload(
	ctx context.Context,
	key string,
	body io.Reader,
	contentType string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        body,
		ContentType: &contentType,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

func (s *S3Service) S3Delete(
	ctx context.Context,
	key string,
) error {

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}

func (s *S3Service) S3Download(
	ctx context.Context,
	key string,
) (io.ReadCloser, error) {

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return result.Body, nil
}

func (s *S3Service) S3GetUrl(
	ctx context.Context,
	key string,
) (string, error) {

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: &s.bucket,
			Key:    &key,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate S3 URL: %w", err)
	}

	return request.URL, nil
}
