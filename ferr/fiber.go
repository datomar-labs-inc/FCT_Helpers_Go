package ferr

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

func Middleware(withStack bool) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		var extractedError *Error

		// Run everything inside an extra function so that a panic can be easily caught
		func() {
			defer func() {
				if err := recover(); err != nil {
					code := http.StatusInternalServerError

					extractedError = &Error{
						Message:  fmt.Sprintf("%+v", err),
						Type:     ETGeneric,
						Code:     CodePanic,
						HTTPCode: &code,
					}
				}
			}()

			err := c.Next()
			if err != nil {
				extractedError = Infer(extractedError)
			}
		}()

		if extractedError != nil {
			if extractedError.HTTPCode == nil {
				code := http.StatusInternalServerError
				extractedError.HTTPCode = &code
			}

			return c.Status(*extractedError.HTTPCode).JSON(extractedError.ToAPIResponseError(withStack))
		}

		return nil
	}
}
