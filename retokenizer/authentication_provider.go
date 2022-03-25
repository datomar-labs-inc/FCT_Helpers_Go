package retokenizer

import (
	"context"
	"errors"
)

type UserInfo struct {
	Sub           string `json:"sub"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Nickname      string `json:"nickname"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	UpdatedAt     string `json:"updated_at"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

var ErrInvalidAuthorization = errors.New("invalid authorization")

type AuthenticationProvider interface {
	GetUserInfo(ctx context.Context, token string) (*UserInfo, error)
}
