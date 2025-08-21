package service

import (
	"archive"
	"archive/pkg/repository"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	// рекомендуется вынести в конфиг приложения
	salt       = "hjqrhjqw124617ajfhajs"
	signingKey = "archive#test#just#some###24214texTTTT#S"
	tokenTTL   = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int64 `json:"user_id"`
}

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(ctx context.Context, user archive.User) (int64, error) {
	if user.Login == "" || user.PasswordHash == "" {
		return 0, errors.New("login and password required")
	}

	// Хешируем пароль перед передачей в репозиторий (если вызываете fn_register_user на уровне БД — адаптируйте)
	user.PasswordHash = s.generatePasswordHash(user.PasswordHash)

	return s.repo.CreateUser(ctx, user)
}

func (s *AuthService) GenerateToken(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("username/password required")
	}
	hashed := s.generatePasswordHash(password)

	u, err := s.repo.GetUser(ctx, username, hashed)
	if err != nil {
		return "", err
	}

	claims := tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId: u.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) ParseToken(ctx context.Context, accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}
	return claims.UserId, nil
}

func (s *AuthService) generatePasswordHash(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum([]byte(salt)))
}
