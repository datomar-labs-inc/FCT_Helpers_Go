package ferr

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func Middleware(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		fctErr := Infer(err)

		if fctErr.HTTPCode == nil {
			code := http.StatusInternalServerError
			fctErr.HTTPCode = &code
		}

		return c.Status(*fctErr.HTTPCode).JSON(fctErr.ToAPIResponseError())
	}

	return nil
}
