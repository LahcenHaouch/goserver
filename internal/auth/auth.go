package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPassword), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now().UTC()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		Subject:   userId.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims

	if _, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(tokenSecret), nil
	}); err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(claims.Subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	headerArr := strings.Split(auth, " ")

	if len(headerArr) != 2 {
		return "", errors.New("no bearer token found")
	}

	return headerArr[1], nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	header := headers.Get("Authorization")
	headerArr := strings.Split(header, " ")

	if len(headerArr) != 2 {
		return "", errors.New("no api key found")
	}

	return headerArr[1], nil
}
