package auth

import "github.com/golang-jwt/jwt/v4"

// GenerateToken, 登录的时候生成AccessToken和RefreshToken
func GenerateToken(secret string, iat, seconds, Id int64) (string, error) {
	claims := jwt.MapClaims{
		"Id":  Id,
		"iat": iat,
		"exp": iat + seconds,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(secret))
}
