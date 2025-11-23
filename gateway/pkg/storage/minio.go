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
	client *minio.Client
	bucket string
}

func NewMinioClient(env bootstrap.Env) (*MinioClient, error) {
	endpoint := fmt.Sprintf("%s:%s", env.Minio.Host, env.Minio.Port)
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(env.Minio.UserRoot, env.Minio.PasswordRoot, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	bucketName := "barbod-docs"
	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err == nil && !exists {
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
	}

	return &MinioClient{
		client: minioClient,
		bucket: bucketName,
	}, err
}

func (m *MinioClient) GeneratePresignedUploadURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = time.Minute * 15
	}

	presignedUrl, err := m.client.PresignedPutObject(ctx, m.bucket, objectName, expiry)

	if err != nil {
		return "", err
	}

	return presignedUrl.String(), nil
}

func (m *MinioClient) GetFileURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = time.Minute * 15
	}

	reqParams := make(url.Values)
	presignedUrl, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, reqParams)

	if err != nil {
		return "", err
	}

	return presignedUrl.String(), nil
}
