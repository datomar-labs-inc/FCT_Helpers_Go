package imguploader

import (
	"bytes"
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

func (m *MinioS3UploaderStorage) Store(key string, details *ImageDetails, reader io.Reader) error {
	_, err := m.client.PutObject(m.bucket, key, reader, int64(details.ConvertedSizeBytes), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioS3UploaderStorage) StoreBytes(key string, details *ImageDetails, file []byte) error {
	_, err := m.client.PutObject(m.bucket, key, bytes.NewReader(file), int64(details.ConvertedSizeBytes), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioS3UploaderStorage) Read(key string) (io.ReadCloser, error) {
	file, err := m.client.GetObject(m.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (m *MinioS3UploaderStorage) ReadBytes(key string) ([]byte, error) {
	reader, err := m.Read(key)
	if err != nil {
		return nil, err
	}

	fileBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}
