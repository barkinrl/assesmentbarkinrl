/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package minio

import (
	"context"
	"log/slog"
	"time"

	"github.com/kazimsarikaya/assesmentbarkinrl/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	bucket string
}

func NewClient() *Client {
	config := config.GetConfig()
	
	// Initialize minio client object.
	minioClient, err := minio.New(config.GetMinioEndpoint(), &minio.Options{
		Creds:  credentials.NewStaticV4(config.GetMinioAccessKey(), config.GetMinioSecretKey(), ""),
		Secure: false,
	})
	if err != nil {
		slog.Error("Error creating MinIO client", "error", err)
		return nil
	}

	return &Client{
		client: minioClient,
		bucket: "default",
	}
}

// CreateBucket creates a new bucket if it doesn't exist
func (c *Client) CreateBucket(ctx context.Context, bucketName string) error {
	exists, err := c.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = c.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

// UploadFile uploads a file to MinIO
func (c *Client) UploadFile(ctx context.Context, bucketName, objectName string, filePath string) error {
	_, err := c.client.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	return err
}

// DownloadFile downloads a file from MinIO
func (c *Client) DownloadFile(ctx context.Context, bucketName, objectName, filePath string) error {
	err := c.client.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	return err
}

// ListFiles lists all files in a bucket
func (c *Client) ListFiles(ctx context.Context, bucketName string) ([]string, error) {
	var files []string
	
	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object.Key)
	}
	
	return files, nil
}

// GetPresignedURL generates a presigned URL for a file
func (c *Client) GetPresignedURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
} 