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

func CreateJWTForUser[U any](key []byte, opts *CreateJWTOpts, sub string, user *U) (string, *JWTClaims[U], error) {
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

	ss, err := token.SignedString(key)
	if err != nil {
		return "", nil, err
	}

	return ss, &claims, nil
}

func ValidateUserJWT[U any](key []byte, token, issuer string) (*U, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims[U]{}, KeyBasedKeyFunc(key), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	if claims, ok := parsedToken.Claims.(*JWTClaims[U]); ok && parsedToken.Valid && claims.Valid() == nil {
		if !claims.VerifyIssuer(issuer, true) {
			return nil, errors.New("invalid issuer")
		} else if !claims.VerifyExpiresAt(time.Now(), true) {
			return nil, errors.New("token expired")
		}

		return claims.User, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
