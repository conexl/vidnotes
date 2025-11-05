package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type JWTManager struct {
	accessSecret    string
	refreshSecret   string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

func NewJWTManager(accessSecret, refreshSecret string, AccessDuration, RefreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:    accessSecret,
		refreshSecret:   refreshSecret,
		AccessDuration:  AccessDuration,
		RefreshDuration: RefreshDuration,
	}
}

func (manager *JWTManager) GenerateAccessToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.AccessDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vidnotes-api",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.accessSecret))
}

func (manager *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.RefreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "vidnotes-api",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.refreshSecret))
}

func (manager *JWTManager) GenerateTokenPair(userID string) (*TokenPair, error) {
	accessToken, err := manager.GenerateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := manager.GenerateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(manager.AccessDuration.Seconds()),
	}, nil
}

func (manager *JWTManager) VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(manager.accessSecret), nil
		},
	)

	if err != nil {
		return manager.handleTokenError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (manager *JWTManager) VerifyRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(manager.refreshSecret), nil
		},
	)

	if err != nil {
		return manager.handleTokenError(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (manager *JWTManager) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := manager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return manager.GenerateTokenPair(claims.UserID)
}

func (manager *JWTManager) handleTokenError(err error) (*Claims, error) {
	var jwtErr *jwt.ValidationError
	if errors.As(err, &jwtErr) {
		if jwtErr.Errors&jwt.ValidationErrorExpired != 0 {
			return nil, ErrExpiredToken
		}
	}
	return nil, ErrInvalidToken
}

func (manager *JWTManager) GetRemainingTime(tokenString string) (float64, error) {
	claims, err := manager.VerifyAccessToken(tokenString)
	if err != nil {
		return 0, err
	}

	expiryTime := claims.ExpiresAt.Time
	remaining := time.Until(expiryTime).Seconds()

	if remaining < 0 {
		return 0, ErrExpiredToken
	}

	return remaining, nil
}

// Middleware для Fiber
func (manager *JWTManager) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Authorization header required",
				"message": "Missing authorization token",
			})
		}

		// Проверяем формат "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Invalid authorization format",
				"message": "Authorization header must be in format: Bearer <token>",
			})
		}

		token := parts[1]
		claims, err := manager.VerifyAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "Invalid or expired token",
				"message": "Please login again",
			})
		}

		// Сохраняем userID в контексте
		c.Locals("userID", claims.UserID)
		c.Locals("userEmail", claims.Email)

		return c.Next()
	}
}
