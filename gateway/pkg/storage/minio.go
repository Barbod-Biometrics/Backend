package storage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client       *minio.Client
	publicClient *minio.Client // Client for generating presigned URLs with public endpoint
	bucket       string
}

func NewMinioClient(env bootstrap.Env) (*MinioClient, error) {
	endpoint := fmt.Sprintf("%s:%s", env.Minio.Host, env.Minio.Port)
	creds := credentials.NewStaticV4(env.Minio.UserRoot, env.Minio.PasswordRoot, "")

	// Main client for internal operations (bucket management, etc.)
	// Always uses HTTP internally within Docker network
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: false, // Internal Docker network doesn't need HTTPS
		Region: "us-east-1",
	})
	if err != nil {
		return nil, err
	}

	// Determine public endpoint - this is what external users will use
	// For development: localhost:9000 (docker port mapping)
	// For production: storage.example.com (behind nginx with SSL)
	publicEndpoint := env.Minio.PublicEndpoint
	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}

	// Create a public client for presigned URL generation
	// Uses HTTPS if PublicSecure is true (for production behind nginx with SSL)
	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  creds,
		Secure: env.Minio.PublicSecure, // true for HTTPS (production), false for HTTP (development)
		Region: "us-east-1",
	})
	if err != nil {
		// Fall back to main client if public endpoint is invalid
		publicClient = minioClient
	}

	bucketName := "barbod-docs"
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err == nil && !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	}

	return &MinioClient{
		client:       minioClient,
		publicClient: publicClient,
		bucket:       bucketName,
	}, err
}

// GeneratePresignedUploadURL creates a presigned URL for uploading files.
// Uses the public client so the signature matches when users upload from outside Docker.
func (m *MinioClient) GeneratePresignedUploadURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = time.Minute * 15
	}

	presignedUrl, err := m.publicClient.PresignedPutObject(ctx, m.bucket, objectName, expiry)
	if err != nil {
		return "", err
	}

	return presignedUrl.String(), nil
}

// GetFileURL creates a presigned URL for downloading files.
// Uses the public client so the signature matches when users download from outside Docker.
func (m *MinioClient) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = time.Minute * 15
	}

	reqParams := make(url.Values)
	presignedUrl, err := m.publicClient.PresignedGetObject(ctx, m.bucket, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}

	return presignedUrl.String(), nil
}
