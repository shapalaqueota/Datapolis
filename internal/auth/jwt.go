package auth

import (
	"Datapolis/internal/models"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"time"
)

var (
	ErrInvalidToken = errors.New("недействительный токен")
	ErrExpiredToken = errors.New("срок действия токена истек")
)

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	TokenID  string `json:"jti"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

func GenerateTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, expiresIn, err := generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpiresIn, err := generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

func generateAccessToken(user *models.User) (string, int64, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	duration, err := time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))
	if err != nil {
		duration = 15 * time.Minute
	}

	expirationTime := time.Now().Add(duration)

	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)

	return tokenString, int64(duration.Seconds()), err
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateRefreshToken(user *models.User) (string, int64, error) {
	secret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

	duration, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_EXPIRES_IN"))
	if err != nil {
		duration = 7 * 24 * time.Hour // По умолчанию 7 дней
	}

	expirationTime := time.Now().Add(duration)

	tokenID, err := generateTokenID()
	if err != nil {
		return "", 0, err
	}

	claims := &RefreshClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	return tokenString, int64(duration.Seconds()), err
}

// ValidateToken проверяет access token
func ValidateToken(tokenString string) (*JWTClaims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// ValidateRefreshToken проверяет refresh token
func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	secret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))

	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
