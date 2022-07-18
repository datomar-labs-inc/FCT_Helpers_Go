package imguploader

import (
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/gofiber/fiber/v2"
	"io"
	"io/ioutil"
	"net/http"
)

func (i *ImageUploader) FiberUploadHandler(c *fiber.Ctx) (*ImageDetails, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	f, err := file.Open()
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	defer f.Close()

	details, err := i.Upload(c.UserContext(), file.Filename, f)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return details, nil
}

func (i *ImageUploader) FiberGetHandler(c *fiber.Ctx, key string) error {
	reader, details, err := i.storage.Read(c.UserContext(), key)
	if err != nil {
		return ferr.Wrap(err)
	}

	if details == nil || details.ConvertedMimeType == "" || details.ConvertedSizeBytes == 0 {
		image, err := ioutil.ReadAll(reader)
		if err != nil {
			return ferr.Wrap(err)
		}

		contentType := http.DetectContentType(image[:512])

		details = &ImageDetails{
			ID:                 key,
			ConvertedSizeBytes: uint64(len(image)),
			ConvertedMimeType:  contentType,
		}
	}

	c.Set("Content-Type", details.ConvertedMimeType)
	c.Set("Content-Length", fmt.Sprintf("%d", details.ConvertedSizeBytes))

	if i.cache {
		c.Set("Cache-Control", "public, max-age=31536000")
	}

	c.Response().SetStatusCode(http.StatusOK)

	_, err = io.Copy(c.Response().BodyWriter(), reader)
	if err != nil {
		lggr.Get("upload-image").Warn("failed to send body")
	}

	return nil
}
