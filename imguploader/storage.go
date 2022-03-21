package imguploader

import (
	"context"
	"io"
)

type ImageUploaderStorage interface {
	Store(ctx context.Context, key string, details *ImageDetails, reader io.Reader) error
	StoreBytes(ctx context.Context, key string, details *ImageDetails, file []byte) error
	Read(ctx context.Context, key string) (io.ReadCloser, *ImageDetails, error)
	ReadBytes(ctx context.Context, key string) ([]byte, error)
}
