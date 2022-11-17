package main

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Token contains the Token itself in addition to related metadata. Use the
// function NewToken to construct a new Token.
type Token struct {
	Timestamp   string `json:"timestamp"`
	Fingerprint string `json:"fingerprint"`
	Secret      string `json:"secret"`
}

// NewToken creates a token representation that includes metadata.
func NewToken(secret string) Token {
	salt := "03c49494-c1f3-4b3c-a9e3-28b1c4e42177"
	return Token{
		Timestamp:   time.Now().Format(time.RFC3339),
		Fingerprint: fmt.Sprintf("%x", sha512.Sum512_256([]byte(salt+secret))),
		Secret:      secret,
	}
}

// ExtractToken returns token value from a given map of headers based on a
// given list of possible token header names. First match is returned. Error
// is returned if no match occurs and fallbackToken is not set.
func ExtractToken(
	headers http.Header,
	tokenHeaderNames []string,
	fallbackToken string,
) (Token, error) {
	var secret string
	var err error = nil

	// Extract secret from header.
	for _, tokenHeaderName := range tokenHeaderNames {
		if tokenHeader, ok := headers[tokenHeaderName]; ok {
			if len(tokenHeader) > 0 && len(tokenHeader[0]) > 0 {
				secret = tokenHeader[0]
				break
			}
		}
	}

	// Remove optional authorization type.
	secret = strings.TrimPrefix(secret, "Bearer ")

	// Use optional fallback token.
	if len(secret) == 0 && len(fallbackToken) > 0 {
		secret = fallbackToken
	}

	if len(secret) == 0 {
		err = errors.New("failed to find token")
	}

	return NewToken(secret), err
}
