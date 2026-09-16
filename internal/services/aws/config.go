package aws

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/textract"
)

// AWSConfig holds all AWS configuration and service clients
type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	S3Bucket        string
	SQSQueueURL     string
	SESFromEmail    string
	DynamoDBTable   string

	Config aws.Config

	S3Client       *s3.Client
	SQSClient      *sqs.Client
	SESClient      *ses.Client
	DynamoDBClient *dynamodb.Client
	TextractClient *textract.Client
}

func InitializeAWSConfig(ctx context.Context) (*AWSConfig, error) {
	log.Println("Initializing AWS configuration...")

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	// S3_BUCKET / SQS_QUEUE_URL (no AWS_ prefix) match the names used in
	// this project's .env and the Python AI worker's .env.
	s3Bucket := os.Getenv("S3_BUCKET")
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	sesFromEmail := os.Getenv("AWS_SES_FROM_EMAIL")
	dynamoDBTable := os.Getenv("AWS_DYNAMODB_TABLE")

	if sesFromEmail == "" {
		sesFromEmail = "noreply@mindmesh.com"
	}
	if dynamoDBTable == "" {
		dynamoDBTable = "mindmesh-data"
	}

	var cfg aws.Config
	var err error

	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	if accessKeyID != "" && secretAccessKey != "" {
		log.Println("Using explicit AWS credentials from environment variables")
		configOptions = append(configOptions, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		))
	} else {
		log.Println("Using default AWS credential chain")
	}

	cfg, err = config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	awsConfig := &AWSConfig{
		Region:          region,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		S3Bucket:        s3Bucket,
		SQSQueueURL:     sqsQueueURL,
		SESFromEmail:    sesFromEmail,
		DynamoDBTable:   dynamoDBTable,
		Config:          cfg,
		S3Client:        s3.NewFromConfig(cfg),
		SQSClient:       sqs.NewFromConfig(cfg),
		SESClient:       ses.NewFromConfig(cfg),
		DynamoDBClient:  dynamodb.NewFromConfig(cfg),
		TextractClient:  textract.NewFromConfig(cfg),
	}

	log.Printf("AWS configuration initialized successfully - Region: %s, S3 Bucket: %s, DynamoDB Table: %s\n", region, s3Bucket, dynamoDBTable)

	return awsConfig, nil
}

func (ac *AWSConfig) GetS3Client() *s3.Client {
	return ac.S3Client
}

func (ac *AWSConfig) GetSQSClient() *sqs.Client {
	return ac.SQSClient
}

func (ac *AWSConfig) GetSESClient() *ses.Client {
	return ac.SESClient
}

func (ac *AWSConfig) GetDynamoDBClient() *dynamodb.Client {
	return ac.DynamoDBClient
}

func (ac *AWSConfig) GetTextractClient() *textract.Client {
	return ac.TextractClient
}
