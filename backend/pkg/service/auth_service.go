package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"archive"
	"archive/pkg/repository"

	"github.com/dgrijalva/jwt-go"
)

const (
	// желательно вынести в конфиг приложения
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

	return s.generateTokenForUserID(u.ID)
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
	if !ok || !token.Valid {
		return 0, errors.New("invalid token claims")
	}
	return claims.UserId, nil
}

// RefreshToken — если переданный токен валиден (подпись и срок жизни),
// создаёт и возвращает новый токен для того же user_id.
func (s *AuthService) RefreshToken(ctx context.Context, accessToken string) (string, error) {
	userId, err := s.ParseToken(ctx, accessToken)
	if err != nil {
		return "", err
	}
	return s.generateTokenForUserID(userId)
}

func (s *AuthService) generateTokenForUserID(userID int64) (string, error) {
	claims := tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) generatePasswordHash(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return fmt.Sprintf("%x", h.Sum([]byte(salt)))
}
