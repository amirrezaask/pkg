package objectstore

import (
	"context"
	"io"
	"time"

	"github.com/amirrezaask/pkg/errors"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	_MINIO_HEALTH_CHECK_AFTER = time.Second * 2
)

type MinioClient struct {
	bucketName string

	c *minio.Client
}

func NewMinio(ctx context.Context, bucketName string, endpoint string, accessID string, secretAccessID string) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Region: "us-east-1",
		Creds:  credentials.NewStaticV4(accessID, secretAccessID, ""),
		Secure: false,
	})
	if err != nil {
		return nil, errors.Wrap(err, "minio client cannot be created")
	}
	_, _ = client.HealthCheck(_MINIO_HEALTH_CHECK_AFTER)

	if !client.IsOnline() {
		return nil, errors.Newf("minio endpoint is offline")
	}
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, errors.Wrap(err, "minio bucket exists failed")
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "cannot make new minio bucket")
		}
	}

	if err != nil {
		return nil, err
	}

	return &MinioClient{c: client, bucketName: bucketName}, nil
}

func (m *MinioClient) Store(ctx context.Context, name string, r io.Reader, size int, expireAt time.Time) (minio.UploadInfo, error) {
	ui, err := m.c.PutObject(ctx, m.bucketName, name, r, int64(size), minio.PutObjectOptions{
		Expires: expireAt,
	})

	if err != nil {
		return minio.UploadInfo{}, err
	}

	return ui, nil
}

func (m *MinioClient) Get(ctx context.Context, name string) (io.Reader, error) {
	obj, err := m.c.GetObject(ctx, m.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}
