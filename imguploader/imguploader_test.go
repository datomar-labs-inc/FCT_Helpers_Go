package imguploader

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestImageUploader_processImageStream(t *testing.T) {
	t.Run("png", testPNG)
	t.Run("jpg", testJpeg)
	t.Run("gif", testGif)
	t.Run("webp", testWebp)
}

func testPNG(t *testing.T) {
	mockStorage := NewMockStorage()
	upl := NewImageUploader(mockStorage)

	f, err := os.OpenFile("./png.png", os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	details, err := upl.Upload(context.Background(), f.Name(), f)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 1280, details.Width)
		assert.Equal(t, 720, details.Height)
		assert.Equal(t, f.Name(), details.OriginalFileName)
		assert.Equal(t, FormatPNG, details.OriginalMimeType)
		assert.Equal(t, mockStorage.BytesStored(), details.ConvertedSizeBytes)
	}
}

func testJpeg(t *testing.T) {
	mockStorage := NewMockStorage()
	upl := NewImageUploader(mockStorage)

	f, err := os.OpenFile("./jpg.jpg", os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	details, err := upl.Upload(context.Background(), f.Name(), f)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 1200, details.Width)
		assert.Equal(t, 1200, details.Height)
		assert.Equal(t, f.Name(), details.OriginalFileName)
		assert.Equal(t, FormatJpeg, details.OriginalMimeType)
		assert.Equal(t, mockStorage.BytesStored(), details.ConvertedSizeBytes)
	}
}

func testGif(t *testing.T) {
	mockStorage := NewMockStorage()
	upl := NewImageUploader(mockStorage)

	f, err := os.OpenFile("./gif.gif", os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	details, err := upl.Upload(context.Background(), f.Name(), f)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 500, details.Width)
		assert.Equal(t, 400, details.Height)
		assert.Equal(t, f.Name(), details.OriginalFileName)
		assert.Equal(t, FormatGif, details.OriginalMimeType)
		assert.Equal(t, mockStorage.BytesStored(), details.ConvertedSizeBytes)
	}
}

func testWebp(t *testing.T) {
	mockStorage := NewMockStorage()
	upl := NewImageUploader(mockStorage)

	f, err := os.OpenFile("./webp.webp", os.O_RDONLY, os.ModePerm)
	if err != nil {
		t.Error(err)
		return
	}

	defer f.Close()

	details, err := upl.Upload(context.Background(), f.Name(), f)
	if err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 550, details.Width)
		assert.Equal(t, 368, details.Height)
		assert.Equal(t, f.Name(), details.OriginalFileName)
		assert.Equal(t, FormatWebp, details.OriginalMimeType)
		assert.Equal(t, mockStorage.BytesStored(), details.ConvertedSizeBytes)
	}
}
