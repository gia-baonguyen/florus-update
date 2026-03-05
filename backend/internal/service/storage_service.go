package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/florus/backend/internal/config"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageService interface {
	UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, error)
	UploadBase64(base64Data string, folder string) (string, error)
	DeleteFile(objectName string) error
	GetFileURL(objectName string) string
}

type storageService struct {
	client     *minio.Client
	bucketName string
	endpoint   string
	useSSL     bool
}

func NewStorageService(cfg *config.MinioConfig) (StorageService, error) {
	// Initialize MinIO client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Create bucket if not exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created MinIO bucket: %s", cfg.Bucket)

		// Set bucket policy to allow public read access
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, cfg.Bucket)
		err = client.SetBucketPolicy(ctx, cfg.Bucket, policy)
		if err != nil {
			log.Printf("Warning: failed to set bucket policy: %v", err)
		}
	}

	return &storageService{
		client:     client,
		bucketName: cfg.Bucket,
		endpoint:   cfg.Endpoint,
		useSSL:     cfg.UseSSL,
	}, nil
}

// UploadFile uploads a multipart file to MinIO
func (s *storageService) UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	ctx := context.Background()

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	objectName := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to MinIO
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.GetFileURL(objectName), nil
}

// UploadBase64 uploads a base64 encoded image to MinIO
func (s *storageService) UploadBase64(base64Data string, folder string) (string, error) {
	ctx := context.Background()

	// Parse base64 data URL
	// Format: data:image/png;base64,iVBORw0KGgo...
	var contentType, base64Content string

	if strings.HasPrefix(base64Data, "data:") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid base64 data URL format")
		}

		// Extract content type
		metaParts := strings.SplitN(parts[0], ";", 2)
		contentType = strings.TrimPrefix(metaParts[0], "data:")
		base64Content = parts[1]
	} else {
		// Plain base64 without data URL prefix
		contentType = "image/jpeg"
		base64Content = base64Data
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Determine file extension from content type
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	}

	// Generate unique filename
	objectName := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	// Upload to MinIO
	reader := strings.NewReader(string(decoded))
	_, err = s.client.PutObject(ctx, s.bucketName, objectName, reader, int64(len(decoded)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.GetFileURL(objectName), nil
}

// DeleteFile deletes a file from MinIO
func (s *storageService) DeleteFile(objectName string) error {
	ctx := context.Background()

	// Extract object name from full URL if needed
	if strings.HasPrefix(objectName, "http") {
		// Parse URL to get object name
		parts := strings.SplitN(objectName, s.bucketName+"/", 2)
		if len(parts) == 2 {
			objectName = parts[1]
		}
	}

	err := s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFileURL returns the public URL for a file
func (s *storageService) GetFileURL(objectName string) string {
	protocol := "http"
	if s.useSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, s.endpoint, s.bucketName, objectName)
}

// GetPresignedURL returns a presigned URL for temporary access (optional, for private buckets)
func (s *storageService) GetPresignedURL(objectName string, expiry time.Duration) (string, error) {
	ctx := context.Background()
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// UploadFromReader uploads from an io.Reader
func (s *storageService) UploadFromReader(reader io.Reader, size int64, objectName string, contentType string) (string, error) {
	ctx := context.Background()

	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return s.GetFileURL(objectName), nil
}
