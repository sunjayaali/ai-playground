package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims is the custom JWT payload. It embeds the standard
// registered claims.
type Claims struct {
	jwt.RegisteredClaims
}

// TokenService signs and verifies the two token kinds. Both are
// HMAC-SHA256 over the same secret; the JWT header "typ" carries
// the kind so an access token can never be replayed as a refresh
// token, or the other way around.
type TokenService struct {
	secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewTokenService(secret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

// Issue mints a fresh access/refresh pair for subject (the
// user's UUID). The pair comes back as an oauth2.Token — the
// same struct golang.org/x/oauth2 hands to its clients — so
// a JSON response of it is the standard token shape.
func (t *TokenService) Issue(subject string) (*oauth2.Token, error) {
	now := time.Now()
	access, expiry, err := t.sign(subject, TokenTypeAccess, t.AccessTTL, now)
	if err != nil {
		return nil, err
	}
	refresh, _, err := t.sign(subject, TokenTypeRefresh, t.RefreshTTL, now)
	if err != nil {
		return nil, err
	}
	return &oauth2.Token{
		AccessToken:  access,
		TokenType:    "Bearer",
		RefreshToken: refresh,
		Expiry:       expiry,
		ExpiresIn:    int64(t.AccessTTL / time.Second),
	}, nil
}

// sign mints one token of the given type and returns it with
// the instant its claims expire.
func (t *TokenService) sign(subject, typ string, ttl time.Duration, now time.Time) (string, time.Time, error) {
	expiry := now.Add(ttl)
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok.Header["typ"] = typ
	signed, err := tok.SignedString(t.secret)
	return signed, expiry, err
}

// Key returns the HMAC secret used to sign tokens.
func (t *TokenService) Key() []byte {
	return t.secret
}

// KeyFunc returns a jwt.Keyfunc that validates the signing
// method (HMAC) and that the token header's "typ" matches
// kind.
func (t *TokenService) KeyFunc(kind string) jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		if token.Header["typ"] != kind {
			return nil, fmt.Errorf("unexpected token type %v", token.Header["typ"])
		}
		return t.secret, nil
	}
}

// Parse validates raw and returns its claims. kind pins which token
// type the caller accepts, so the wrong kind fails here.
func (t *TokenService) Parse(raw, kind string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, t.KeyFunc(kind))
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
