package flagger

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	"github.com/gofiber/fiber/v2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type WebhookSender struct {
	key         []byte
	url         string
	client      *http.Client
	hasInternal bool
	internal    *fiber.App
}

func NewSender(key string, url string) *WebhookSender {
	return &WebhookSender{
		key: []byte(key),
		url: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// WithInternal will set an internal fiber instance (so that flagger may call it's self)
// When the webhook url begins with internal:// this endpoint will be used
func (ws *WebhookSender) WithInternal(app *fiber.App) *WebhookSender {
	ws.hasInternal = true
	ws.internal = app
	return ws
}

func (ws *WebhookSender) send(url string, webhook *Webhook) ([]byte, error) {
	jsonBytes, err := json.Marshal(webhook)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	hash := hmac.New(sha1.New, ws.key)
	hash.Write(jsonBytes)

	hashString := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequest("POST", strings.TrimPrefix(url, "internal://"), bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Flagger-Sig", hashString)

	var res *http.Response

	if strings.HasPrefix(url, "internal://") {
		res, err = ws.internal.Test(req)
		if err != nil {
			return nil, ferr.Wrap(err)
		}
	} else {
		res, err = ws.client.Do(req)
		if err != nil {
			return nil, ferr.Wrap(err)
		}
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed request (%d): %s", res.StatusCode, string(body))
	}

	return body, nil
}

func (ws *WebhookSender) GetUserInfo(req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	body, err := ws.send(ws.url, &Webhook{
		Op:          OpGetUserInfo,
		GetUserInfo: req,
	})
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	var res GetUserInfoResponse

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return &res, nil
}

func (ws *WebhookSender) SearchUsers(req *SearchUsersRequest) (*SearchUsersResponse, error) {
	body, err := ws.send(ws.url, &Webhook{
		Op:          OpSearchUsers,
		SearchUsers: req,
	})
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	var res SearchUsersResponse

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return &res, nil
}
