package imguploader

import (
	"bytes"
	"github.com/miolini/datacounter"
	"io"
)

type MockStorage struct {
	countedWriter *datacounter.WriterCounter
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		countedWriter: datacounter.NewWriterCounter(io.Discard),
	}
}

func (m *MockStorage) Store(key string, details *ImageDetails, reader io.Reader) error {
	_, err := io.Copy(m.countedWriter, reader)
	if err != nil {
		return err
	}

	return nil
}

func (m *MockStorage) StoreBytes(key string, details *ImageDetails, file []byte) error {
	return nil
}

func (m *MockStorage) Read(key string) (io.ReadCloser, *ImageDetails, error) {
	return io.NopCloser(bytes.NewReader([]byte("hi there"))), nil, nil
}

func (m *MockStorage) ReadBytes(key string) ([]byte, error) {
	return []byte("hi there"), nil
}

func (m *MockStorage) BytesStored() uint64 {
	return m.countedWriter.Count()
}
