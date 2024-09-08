package jwtTool

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

var (
	TokenExpired     error  = errors.New("Token is expired")
	TokenNotValidYet error  = errors.New("Token not active yet")
	TokenMalformed   error  = errors.New("That's not even a token")
	TokenInvalid     error  = errors.New("Couldn't handle this token:")
	SignKey          string = "test"
)

func CreateToken(claims jwt.RegisteredClaims, key []byte) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}

func ParseToken(tokenStr string, key []byte) (claims jwt.RegisteredClaims, err error) {
	token, err := jwt.ParseWithClaims(
		tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return key, nil
		},
	)
	if err != nil {
		return
	}
	if !token.Valid {
		err = errors.New("parse token fail")
	}
	return
}
