package imguploader

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v6"
	"io"
	"io/ioutil"
)

type MinioS3UploaderStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioS3Storage(client *minio.Client, bucket string) *MinioS3UploaderStorage {
	return &MinioS3UploaderStorage{
		client: client,
		bucket: bucket,
	}
}

func (m *MinioS3UploaderStorage) Store(ctx context.Context, key string, details *ImageDetails, reader io.Reader) error {
	_, err := m.client.PutObjectWithContext(ctx, m.bucket, key, reader, int64(details.ConvertedSizeBytes), minio.PutObjectOptions{
		ContentType: details.ConvertedMimeType,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioS3UploaderStorage) StoreBytes(ctx context.Context, key string, details *ImageDetails, file []byte) error {
	_, err := m.client.PutObjectWithContext(ctx, m.bucket, key, bytes.NewReader(file), int64(details.ConvertedSizeBytes), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioS3UploaderStorage) Read(ctx context.Context, key string) (io.ReadCloser, *ImageDetails, error) {
	file, err := m.client.GetObjectWithContext(ctx, m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	details := &ImageDetails{
		ID:                 key,
		ConvertedMimeType:  stat.ContentType,
		ConvertedSizeBytes: uint64(stat.Size),
	}

	return file, details, nil
}

func (m *MinioS3UploaderStorage) ReadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, _, err := m.Read(ctx, key)
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
