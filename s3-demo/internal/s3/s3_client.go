package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	bucket string
	client *s3.Client
	region string
	creds  aws.CredentialsProvider
}

func NewClient(accessKeyID, secretAccessKey, region, bucket string) (*Client, error) {
	credsProvider := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")

	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		bucket: bucket,
		client: s3.NewFromConfig(cfg),
		region: region,
		creds:  credsProvider,
	}, nil
}

// GeneratePresignedUploadURL creates a pre-signed PUT URL with Content-Type signed.
// Uses low-level v4.Signer directly to ensure Content-Type is included in X-Amz-SignedHeaders.
func (c *Client) GeneratePresignedUploadURL(key, contentType string) (string, error) {
	creds, err := c.creds.Retrieve(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to retrieve credentials: %w", err)
	}

	endpoint := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.bucket, c.region, key)

	req, err := http.NewRequest(http.MethodPut, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// MUST set Content-Type header so it's included in signed headers
	// This ensures the browser can send the exact same Content-Type
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	req.Header.Set("Content-Type", contentType)

	q := req.URL.Query()
	q.Set("X-Amz-Expires", strconv.FormatInt(int64(5*time.Minute/time.Second), 10))
	req.URL.RawQuery = q.Encode()

	signer := v4.NewSigner()
	presignedURL, _, err := signer.PresignHTTP(
		context.Background(),
		creds,
		req,
		"UNSIGNED-PAYLOAD",
		"s3",
		c.region,
		time.Now().UTC(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}

	return presignedURL, nil
}

func (c *Client) UploadObject(ctx context.Context, key, contentType string, body io.Reader) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}
