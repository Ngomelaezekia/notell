package main

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMakeSlug(t *testing.T) {
	cases := map[string]string{
		"My Channel": "my-channel",
		"  News & Sports  ": "news-sports",
		"Hello---World": "hello-world",
	}
	for input, want := range cases {
		if got := makeSlug(input); got != want { t.Fatalf("makeSlug(%q) = %q, want %q", input, got, want) }
	}
}

func TestChannelTypes(t *testing.T) {
	if ChannelPublic == ChannelPrivate { t.Fatal("public and private channel types must differ") }
	if ChannelActive == ChannelDisabled { t.Fatal("active and disabled statuses must differ") }
}

func TestClaimsIssuerAndExpiry(t *testing.T) {
	claims := Claims{UserID: 42, RegisteredClaims: jwt.RegisteredClaims{Issuer: "notell-api", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	if claims.UserID != 42 || claims.Issuer != "notell-api" || claims.ExpiresAt == nil { t.Fatal("invalid channel auth contract claims") }
}
