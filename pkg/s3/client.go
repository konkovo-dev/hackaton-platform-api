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
	client *minio.Client
	config *Config
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

	return &Client{
		client: minioClient,
		config: config,
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

	return nil
}

func (c *Client) GeneratePresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, bucket, key, expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	urlStr := presignedURL.String()
	
	if c.config.PublicEndpoint != c.config.Endpoint {
		urlStr = c.replaceEndpoint(urlStr, c.config.Endpoint, c.config.PublicEndpoint)
	}

	return urlStr, nil
}

func (c *Client) GeneratePresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := c.client.PresignedGetObject(ctx, bucket, key, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	urlStr := presignedURL.String()
	
	if c.config.PublicEndpoint != c.config.Endpoint {
		urlStr = c.replaceEndpoint(urlStr, c.config.Endpoint, c.config.PublicEndpoint)
	}

	return urlStr, nil
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

func (c *Client) replaceEndpoint(urlStr, oldEndpoint, newEndpoint string) string {
	scheme := "http://"
	if c.config.UseSSL {
		scheme = "https://"
	}
	
	oldURL := scheme + oldEndpoint
	newURL := scheme + newEndpoint
	
	return strings.Replace(urlStr, oldURL, newURL, 1)
}
