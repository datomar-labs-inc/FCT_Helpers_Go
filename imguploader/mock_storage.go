package imguploader

import (
	"bytes"
	"context"
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

func (m *MockStorage) Store(_ context.Context, _ string, _ *ImageDetails, reader io.Reader) error {
	_, err := io.Copy(m.countedWriter, reader)
	if err != nil {
		return ferr.Wrap(err)
	}

	return nil
}

func (m *MockStorage) StoreBytes(_ context.Context, _ string, _ *ImageDetails, _ []byte) error {
	return nil
}

func (m *MockStorage) Read(_ context.Context, _ string) (io.ReadCloser, *ImageDetails, error) {
	return io.NopCloser(bytes.NewReader([]byte("hi there"))), nil, nil
}

func (m *MockStorage) ReadBytes(_ context.Context, _ string) ([]byte, error) {
	return []byte("hi there"), nil
}

func (m *MockStorage) BytesStored() uint64 {
	return m.countedWriter.Count()
}
