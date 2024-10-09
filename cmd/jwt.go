package cmd

import (
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type SessionClaims struct {
	RoleARN string `json:"role_arn"`
	//The issuer of the initial OIDC refresh token
	IIssuer string `json:"initial_issuer"`
	jwt.RegisteredClaims
}

func createRS256PolicyToken(issuer, iIssuer, subject, roleARN string, expiry time.Duration) (*jwt.Token) {
	claims := &SessionClaims{
		roleARN,
		iIssuer,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    issuer,
			Subject:   subject,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token
}

func CreateSignedToken(t *jwt.Token, signingKey *rsa.PrivateKey) (string, error) {
	tokenStr, err := t.SignedString(signingKey)
	return tokenStr, err
}

// ExtractOIDCTokenClaims extracts JWT claims from a security token using the public key of the
// OIDC provider if the OIDC provider is registered key
func ExtractOIDCTokenClaims(token string) (*SessionClaims, error) {
	return ExtractTokenClaims(token, oidcKeyFunc)
}


// ExtractTokenClaims extracts JWT claims using a key functions
func ExtractTokenClaims(token string, keyFunc func (t *jwt.Token) (interface{}, error)) (*SessionClaims, error) {
	if token == "" {
		return nil, errors.New("invalid argument")
	}

	policyClaims := SessionClaims{}

	if _, err := jwt.ParseWithClaims(token, &policyClaims, keyFunc); err != nil {
		return nil, err
	}

	return &policyClaims, nil
}