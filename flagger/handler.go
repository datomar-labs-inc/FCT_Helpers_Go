package flagger

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type WebhookHandler struct {
	webhooker WebHooker
	key       []byte
}

func NewHandler(wh WebHooker, key string) *WebhookHandler {
	return &WebhookHandler{webhooker: wh, key: []byte(key)}
}

func (wh *WebhookHandler) processBody(ctx context.Context, body []byte, signature string) ([]byte, error) {
	valid := wh.verifyBody(body, signature)

	if !valid {
		return nil, errors.New("invalid signature")
	}

	var webhook Webhook

	err := json.Unmarshal(body, &webhook)
	if err != nil {
		return nil, err
	}

	var result any

	switch webhook.Op {
	case OpGetUserInfo:
		result, err = wh.webhooker.GetUserInfo(ctx, webhook.GetUserInfo)
		if err != nil {
			return nil, err
		}
	case OpSearchUsers:
		result, err = wh.webhooker.SearchUsers(ctx, webhook.SearchUsers)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid op " + webhook.Op)
	}

	return json.Marshal(result)
}

func (wh *WebhookHandler) FiberMiddleware(c *fiber.Ctx) error {
	signature := c.Get("X-Flagger-Sig")

	body := c.Body()

	res, err := wh.processBody(c.UserContext(), body, signature)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Type", "application/json")
	c.Set("Content-Length", fmt.Sprintf("%d", len(res)))

	return c.Status(http.StatusOK).Send(res)
}

func (wh *WebhookHandler) verifyBody(body []byte, expectedHash string) bool {
	hash := hmac.New(sha1.New, wh.key)
	hash.Write(body)

	hashString := hex.EncodeToString(hash.Sum(nil))

	return hashString == expectedHash
}
