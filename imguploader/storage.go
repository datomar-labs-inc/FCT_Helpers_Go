package imguploader

import "io"

type ImageUploaderStorage interface {
	Store(key string, details *ImageDetails, reader io.Reader) error
	StoreBytes(key string, details *ImageDetails, file []byte) error
	Read(key string) (io.ReadCloser, error)
	ReadBytes(key string) ([]byte, error)
}
