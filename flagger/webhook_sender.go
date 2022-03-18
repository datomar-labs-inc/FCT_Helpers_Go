package flagger

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type WebhookSender struct {
	key    []byte
	url    string
	client *http.Client
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

func (ws *WebhookSender) send(url string, webhook *Webhook) ([]byte, error) {
	jsonBytes, err := json.Marshal(webhook)
	if err != nil {
		return nil, err
	}

	hash := hmac.New(sha1.New, ws.key)
	hash.Write(jsonBytes)

	hashString := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Flagger-Sig", hashString)

	res, err := ws.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("failed request (%d): %s", res.StatusCode, string(body)))
	}

	return body, nil
}

func (ws *WebhookSender) GetUserInfo(req *GetUserInfoRequest) (*GetUserInfoResponse, error) {
	body, err := ws.send(ws.url, &Webhook{
		Op:          OpGetUserInfo,
		GetUserInfo: req,
	})
	if err != nil {
		return nil, err
	}

	var res GetUserInfoResponse

	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
