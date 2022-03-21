package imguploader

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/miolini/datacounter"
	"golang.org/x/image/webp"
	"image"
	"image/png"
	"io"
	"net/http"
)

const (
	FormatPNG  = "image/png"
	FormatJpeg = "image/jpeg"
	FormatGif  = "image/gif"
	FormatWebp = "image/webp"
)

var mimeList = []string{FormatPNG, FormatJpeg, FormatGif, FormatWebp}

type ImageDetails struct {
	ID string `json:"id"`

	Width             int    `json:"width"`
	Height            int    `json:"height"`
	OriginalMimeType  string `json:"original_mime_type"`
	OriginalSizeBytes uint64 `json:"original_size_bytes"`
	OriginalFileName  string `json:"original_file_name"`

	ConvertedMimeType  string `json:"converted_mime_type"`
	ConvertedSizeBytes uint64 `json:"converted_size_bytes"`
}

type ImageUploader struct {
	storage ImageUploaderStorage
	cache   bool
}

func NewImageUploader(storage ImageUploaderStorage) *ImageUploader {
	return &ImageUploader{
		storage: storage,
		cache:   true,
	}
}

func (i *ImageUploader) Upload(name string, reader io.Reader) (*ImageDetails, error) {
	img, details, err := i.decodeImageStream(reader)
	if err != nil {
		return nil, err
	}

	details.OriginalFileName = name
	details.ID = uuid.NewString()

	var imageBuffer bytes.Buffer

	err = png.Encode(&imageBuffer, img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	imaging.Encode(&imageBuffer, img, imaging.PNG)

	details.ConvertedMimeType = FormatPNG
	details.ConvertedSizeBytes = uint64(imageBuffer.Len())

	err = i.storage.Store(details.ID, details, &imageBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to store image: %w", err)
	}

	return details, nil
}

func (i *ImageUploader) decodeImageStream(reader io.Reader) (image.Image, *ImageDetails, error) {
	sniffBuffer := make([]byte, 512)

	// Read the first 512 bytes of the reader
	_, err := reader.Read(sniffBuffer)
	if err != nil {
		return nil, nil, err
	}

	// Attempt to detect the image type
	mime := http.DetectContentType(sniffBuffer)

	validMime := false

	for _, allowedMime := range mimeList {
		if mime == allowedMime {
			validMime = true
			break
		}
	}

	if !validMime {
		return nil, nil, NewInvalidFormatError(mime)
	}

	details := ImageDetails{
		OriginalMimeType: mime,
	}

	// Reconstruct the whole reader stream
	fullImageStream := io.MultiReader(bytes.NewReader(sniffBuffer), reader)
	imageSizeCounter := datacounter.NewReaderCounter(fullImageStream)

	var img image.Image

	switch mime {
	case FormatPNG:
		img, err = imaging.Decode(imageSizeCounter, imaging.AutoOrientation(true))
		if err != nil {
			return nil, nil, err
		}
	case FormatJpeg:
		img, err = imaging.Decode(imageSizeCounter, imaging.AutoOrientation(true))
		if err != nil {
			return nil, nil, err
		}
	case FormatGif:
		img, err = imaging.Decode(imageSizeCounter, imaging.AutoOrientation(true))
		if err != nil {
			return nil, nil, err
		}
	case FormatWebp:
		img, err = webp.Decode(imageSizeCounter)
		if err != nil {
			return nil, nil, err
		}
	default:
		return nil, nil, NewUnsupportedFormatError(mime)
	}

	details.OriginalSizeBytes = imageSizeCounter.Count()
	details.Width = img.Bounds().Size().X
	details.Height = img.Bounds().Size().Y

	return img, &details, nil
}
