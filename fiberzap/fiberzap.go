package fiberzap

import (
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	SkipPaths []string
}

func New(config *Config, logger *lggr.LogWrapper) func(c *fiber.Ctx) error {
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *fiber.Ctx) error {

		// Don't log if this path is skipped
		if _, ok := skipPaths[c.Path()]; ok {
			return c.Next()
		}

		start := time.Now()
		path := c.Path()
		query := string(c.Request().URI().QueryString())

		err := c.Next()

		var fields []zapcore.Field

		if err != nil {
			fields = append(fields, zap.Error(err))
		}

		latency := time.Now().Sub(start)

		fields = append(fields,
			zap.Int("status", c.Response().StatusCode()),
			zap.String("method", c.Method()),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.GetReqHeaders()["User-Agent"]),
			zap.Time("start_time", start),
			zap.Duration("latency", latency),
		)

		logger.Info(path, fields...)

		return nil
	}
}

func Recovery(logger *lggr.LogWrapper) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					logger.Error(c.Path(),
						zap.Any("error", err),
						zap.String("request", c.Request().String()),
					)
					return
				}

				logger.Error("[Recovery from panic]",
					zap.Time("time", time.Now()),
					zap.Any("error", err),
					zap.String("request", c.Request().String()),
				)

				c.
					Status(http.StatusInternalServerError).
					JSON(map[string]interface{}{
						"error": "an unrecoverable error occured",
					})
			}
		}()

		return c.Next()
	}
}