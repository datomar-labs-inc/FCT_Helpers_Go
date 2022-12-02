package lggr

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Middleware(logger *LogWrapper) func(c *fiber.Ctx) error {
	return func (c *fiber.Ctx) error {
		requestID := uuid.NewString()
		lg := logger.With(zap.String("request_id", requestID))
		c.SetUserContext(context.WithValue(c.UserContext(), ContextKey, lg))

		return c.Next()
	}
}


