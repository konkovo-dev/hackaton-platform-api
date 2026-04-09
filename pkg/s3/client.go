package s3

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client        *minio.Client
	presignClient *minio.Client // uses PublicEndpoint for correct presigned URL signatures
	config        *Config
}

func (c *Client) Config() *Config {
	return c.config
}

func NewClient(config *Config) (*Client, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	presignEndpoint := config.PublicEndpoint
	if presignEndpoint == "" {
		presignEndpoint = config.Endpoint
	}
	presignClient, err := minio.New(presignEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create presign minio client: %w", err)
	}

	return &Client{
		client:        minioClient,
		presignClient: presignClient,
		config:        config,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
			Region: c.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	if c.config.PublicRead {
		if err := c.setPublicReadPolicy(ctx, bucketName); err != nil {
			return fmt.Errorf("failed to set public read policy: %w", err)
		}
	}

	return nil
}

// setPublicReadPolicy sets an S3 bucket policy that allows anonymous GET on all objects.
func (c *Client) setPublicReadPolicy(ctx context.Context, bucketName string) error {
	policy := strings.NewReplacer("BUCKET", bucketName).Replace(`{
		"Version":"2012-10-17",
		"Statement":[{
			"Effect":"Allow",
			"Principal":{"AWS":["*"]},
			"Action":["s3:GetObject"],
			"Resource":["arn:aws:s3:::BUCKET/*"]
		}]
	}`)
	return c.client.SetBucketPolicy(ctx, bucketName, policy)
}

func (c *Client) GeneratePresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignedURL, err := c.presignClient.PresignedPutObject(ctx, bucket, key, expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (c *Client) GeneratePresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.presignClient.PresignedGetObject(ctx, bucket, key, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}
	return presignedURL.String(), nil
}

func (c *Client) HeadObject(ctx context.Context, bucket, key string) (size int64, exists bool, err error) {
	objInfo, err := c.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" || errResponse.StatusCode == 404 {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to stat object: %w", err)
	}

	return objInfo.Size, true, nil
}
