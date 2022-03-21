package imguploader

import (
	"fmt"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
)

func (i *ImageUploader) FiberUploadHandler(c *fiber.Ctx) (*ImageDetails, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}

	f, err := file.Open()
	if err != nil {
		return nil, err
	}

	defer f.Close()

	details, err := i.Upload(file.Filename, f)
	if err != nil {
		return nil, err
	}

	return details, nil
}

func (i *ImageUploader) FiberGetHandler(c *fiber.Ctx, details *ImageDetails, key string) error {
	reader, err := i.storage.Read(key)
	if err != nil {
		return err
	}

	c.Set("Content-Type", details.ConvertedMimeType)
	c.Set("Content-Length", fmt.Sprintf("%d", details.ConvertedSizeBytes))
	c.Response().SetStatusCode(http.StatusOK)

	_, err = io.Copy(c.Response().BodyWriter(), reader)
	if err != nil {
		lggr.Get("upload-image").Warn("failed to send body")
	}

	return nil
}
