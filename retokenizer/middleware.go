package retokenizer

import (
	"context"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/cache"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type ContextKey string

var ErrMissingTokenMsg = fiber.Map{
	"error": "missing token",
}

var ErrMalformedTokenMsg = fiber.Map{
	"error": "malformed token",
}

var ErrInvalidAuthorizationMsg = fiber.Map{
	"error": "invalid authorization",
}

var ErrInternalServerErrMsg = fiber.Map{
	"error": "something went wrong",
}

var UserInfoContextKey = ContextKey("__retokenizer_user_info_context_key")
var DefaultAuthCacheDuration = 10 * time.Minute

type MiddlewareOptions struct {
	IgnoreMissingTokens bool
	IgnoreInvalidTokens bool
	Cache               cache.Cache
	AuthProvider        AuthenticationProvider
	AuthCacheDuration   *time.Duration
	Logger              *lggr.LogWrapper
}

func (rt *ReTokenizer) MakeFiberMiddleware(opts *MiddlewareOptions) func(c *fiber.Ctx) error {
	if opts.Cache == nil {
		opts.Cache = cache.NewInMemoryCache()
	}

	if opts.AuthProvider == nil {
		panic("missing auth provider")
	}

	if opts.AuthCacheDuration == nil {
		opts.AuthCacheDuration = &DefaultAuthCacheDuration
	}

	var logger *lggr.LogWrapper

	if opts.Logger != nil {
		logger = opts.Logger
	} else {
		logger = lggr.Get("retokenizer-middleware")
	}

	return func(c *fiber.Ctx) error {
		spanCtx, span := rt.tracer.Start(c.UserContext(), "retokenizer")

		defer func() {
			if span.IsRecording() {
				span.End()
			}
		}()

		token := c.GetReqHeaders()["Authorization"]

		if token == "" || token == "null" {
			if opts.IgnoreMissingTokens {
				return c.Next()
			}

			return c.
				Status(http.StatusUnauthorized).
				JSON(ErrMissingTokenMsg)
		}

		if !strings.HasPrefix(token, "Bearer ") {
			if opts.IgnoreInvalidTokens {
				return c.Next()
			}

			return c.
				Status(http.StatusUnauthorized).
				JSON(ErrMalformedTokenMsg)
		}

		token = strings.TrimSpace(strings.TrimPrefix(token, "Bearer"))

		if token == "" {
			if opts.IgnoreInvalidTokens {
				return c.Next()
			}

			return c.
				Status(http.StatusUnauthorized).
				JSON(ErrMalformedTokenMsg)
		}

		userInfo, err := cache.GetCachedJSONValueWithExpiry(
			spanCtx,
			opts.Cache,
			fmt.Sprintf("auth:%s", token),
			func(ctx context.Context) (*UserInfo, error) {
				return opts.AuthProvider.GetUserInfo(ctx, token)
			},
			opts.AuthCacheDuration,
		)
		if err != nil && err != ErrInvalidAuthorization {
			logger.Error("failed to authorize user", zap.Error(err))

			return c.
				Status(http.StatusInternalServerError).
				JSON(ErrInternalServerErrMsg)
		} else if err == ErrInvalidAuthorization {
			return c.
				Status(http.StatusUnauthorized).
				JSON(ErrInvalidAuthorizationMsg)
		}

		c.SetUserContext(context.WithValue(c.UserContext(), UserInfoContextKey, userInfo))

		return c.Next()
	}
}

// ExtractUserInfo will extract user info attached during auth middleware
// meant to be called with fiber's UserContext
// ie. ExtractUserInfo(c.UserContext())
func ExtractUserInfo(ctx context.Context) *UserInfo {
	if v, ok := ctx.Value(UserInfoContextKey).(*UserInfo); ok {
		return v
	}

	return nil
}
