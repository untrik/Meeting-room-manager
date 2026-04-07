package auth

import (
	"fmt"
	"time"

	"github.com/avito-internships/test-backend-1-untrik/internal/config"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserId string
	Role   string
}

type JWTService struct {
	secret string
}

func NewJWTService(cfg config.Config) *JWTService {
	return &JWTService{
		secret: cfg.JWTSecret,
	}
}

func (s *JWTService) GenerateToken(duration time.Duration, userId string, role string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userId,
		"role":   role,
		"exp":    time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}
func (s *JWTService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	userId, ok := mapClaims["userId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid userId claim")
	}

	role, ok := mapClaims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid role claim")
	}

	return &Claims{
		UserId: userId,
		Role:   role,
	}, nil
}
