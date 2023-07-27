package jwtutil

import "github.com/golang-jwt/jwt/v5"

type JwtUtil struct {
	method jwt.SigningMethod
	key    any
}

func New(method jwt.SigningMethod, key any) *JwtUtil {
	return &JwtUtil{
		method: method,
		key:    key,
	}
}

func (u *JwtUtil) SignClaims(claims jwt.Claims) (string, error) {
	return jwt.NewWithClaims(u.method, claims).SignedString(u.key)
}

func (u *JwtUtil) ParseToClaims(tokenString string, claims jwt.Claims) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return u.key, nil
	})
}
