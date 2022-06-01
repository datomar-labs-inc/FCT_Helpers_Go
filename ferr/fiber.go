package ferr

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func Middleware(withStack bool) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			fctErr := Infer(err)

			if fctErr.HTTPCode == nil {
				code := http.StatusInternalServerError
				fctErr.HTTPCode = &code
			}

			return c.Status(*fctErr.HTTPCode).JSON(fctErr.ToAPIResponseError(withStack))
		}

		return nil
	}
}