package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSService struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSService(client *sqs.Client, queueURL string) *SQSService {
	return &SQSService{
		client:   client,
		queueURL: queueURL,
	}
}

type FileProcessingJob struct {
	JobType   string `json:"job_type"`
	FileID    string `json:"file_id"`
	ProjectID string `json:"project_id"`
	Bucket    string `json:"s3_bucket"`
	Key       string `json:"s3_key"`
}

func (s *SQSService) SendFileProcessingJob(
	ctx context.Context,
	job FileProcessingJob,
) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal file processing job: %w", err)
	}

	messageBody := string(body)
	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.queueURL,
		MessageBody: &messageBody,
	})
	if err != nil {
		return fmt.Errorf("send file processing job to SQS: %w", err)
	}

	return nil
}
