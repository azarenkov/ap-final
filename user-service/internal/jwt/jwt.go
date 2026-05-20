package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

type Claims struct {
	jwtv5.RegisteredClaims
}

func (i *Issuer) Issue(userID string) (token string, jti string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(i.ttl).UTC()
	jti = uuid.NewString()
	c := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, Claims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID,
			ID:        jti,
			IssuedAt:  jwtv5.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	})
	signed, err := c.SignedString(i.secret)
	return signed, jti, expiresAt, err
}

func (i *Issuer) Parse(token string) (subject, jti string, expiresAt time.Time, err error) {
	parsed, err := jwtv5.ParseWithClaims(token, &Claims{}, func(t *jwtv5.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return i.secret, nil
	})
	if err != nil || !parsed.Valid {
		return "", "", time.Time{}, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok {
		return "", "", time.Time{}, errors.New("bad claims type")
	}
	if c.ExpiresAt != nil {
		expiresAt = c.ExpiresAt.Time
	}
	return c.Subject, c.ID, expiresAt, nil
}
