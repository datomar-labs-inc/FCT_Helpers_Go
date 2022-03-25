package retokenizer

import (
	"context"
	"fmt"
	"time"
)

// PlaceholderAuthenticationProvider will take any string as a token, and generate a random user
type PlaceholderAuthenticationProvider struct{}

func (p *PlaceholderAuthenticationProvider) GetUserInfo(_ context.Context, token string) (*UserInfo, error) {
	return &UserInfo{
		Sub:           token,
		GivenName:     "Placeholder-" + token,
		FamilyName:    fmt.Sprintf("Placeholder Family (%s)", token),
		Nickname:      "mr-placeholder-" + token,
		Name:          "p-lace-holder",
		Picture:       fmt.Sprintf("https://i.pravatar.cc/300?u=%s", token),
		Locale:        "en",
		UpdatedAt:     time.Now().Format(time.RFC3339),
		Email:         fmt.Sprintf("brownie.points.dev+%s@datomar.com", token),
		EmailVerified: true,
	}, nil
}
