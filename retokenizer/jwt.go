package retokenizer

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JWTClaims[U any] struct {
	jwt.RegisteredClaims
	User *U `json:"user"`
}

type CreateJWTOpts struct {
	Audience  string
	Issuer    string
	ExpiresIn time.Duration
}

func CreateJWTForUser[U any](rt *ReTokenizer, opts *CreateJWTOpts, sub string, user *U) (string, *JWTClaims[U], error) {
	claims := JWTClaims[U]{
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{opts.Audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(opts.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    opts.Issuer,
			Subject:   sub,
		},
		User: user,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(rt.key)
	if err != nil {
		return "", nil, err
	}

	return ss, &claims, nil
}

func ValidateUserJWT[U any](rt *ReTokenizer, token string) (*U, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims[U]{}, rt.KeyFunc())
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(*JWTClaims[U]); ok && parsedToken.Valid && claims.Valid() == nil {
		return claims.User, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
