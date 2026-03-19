package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	bucket string
	client *s3.Client
	presignClient *s3.PresignClient
}

func NewClient(accessKeyID, secretAccessKey, region, bucket string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &Client{
		bucket:        bucket,
		client:        client,
		presignClient: s3.NewPresignClient(client),
	}, nil
}

func (c *Client) GeneratePresignedUploadURL(key, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	presignedReq, err := c.presignClient.PresignPutObject(
		context.Background(),
		input,
		s3.WithPresignExpires(5*time.Minute),
	)
	if err != nil {
		return "", fmt.Errorf("failed to presign put object: %w", err)
	}

	return presignedReq.URL, nil
}
