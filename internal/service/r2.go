package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/krtw00/konbu/internal/config"
)

const maxAttachmentSize = 5 << 20 // 5MB

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/svg+xml": true,
}

type R2Service struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

func NewR2Service(cfg *config.Config) *R2Service {
	if cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		return nil
	}

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.R2Endpoint),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
	})

	return &R2Service{
		client:    client,
		bucket:    cfg.R2Bucket,
		publicURL: cfg.R2PublicURL,
	}
}

func (s *R2Service) Upload(ctx context.Context, userID uuid.UUID, filename string, body io.Reader, size int64) (string, error) {
	if size > maxAttachmentSize {
		return "", fmt.Errorf("file too large (max %dMB)", maxAttachmentSize>>20)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !allowedImageTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	key := fmt.Sprintf("%s/%s%s", userID.String(), uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload to R2: %w", err)
	}

	return key, nil
}

func (s *R2Service) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("get from R2: %w", err)
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}

	return out.Body, contentType, nil
}

func (s *R2Service) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete from R2: %w", err)
	}
	return nil
}

func (s *R2Service) PublicURL() string {
	return s.publicURL
}
