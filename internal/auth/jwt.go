package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sharon-xa/high-api/internal/utils"
)

func GenerateToken(tokenSecret string, expInMin int) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(
			time.Now().Add(time.Minute * time.Duration(expInMin)),
		),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenSecret))
}

func GenerateAccessToken(userId, userRole, accessTokenSecret string, expInMin int) (string, error) {
	expirationTime := time.Now().Add((time.Minute * time.Duration(expInMin)))

	claims := &AccessClaims{
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessTokenSecret))
}

func GenerateRefreshToken(userId, refreshTokenSecret string, expInDays int) (string, error) {
	expirationTime := time.Now().Add((time.Hour * 24) * time.Duration(expInDays))

	claims := jwt.RegisteredClaims{
		Subject:   userId,
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(refreshTokenSecret))
}

func ParseAccessToken(token, accessTokenSecret string) (claims *AccessClaims, err error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&AccessClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(accessTokenSecret), nil
		},
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, jwt.ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenInvalidClaims) {
			return nil, jwt.ErrTokenInvalidClaims
		}
		return nil, utils.ErrParsingToken
	}

	if !parsedToken.Valid {
		return nil, utils.ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(*AccessClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
