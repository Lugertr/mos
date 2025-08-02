package service

import (
	"center"
	"center/pkg/repository"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	salt       = "hjqrhjqw124617ajfhajs"
	signingKey = "center#test#just#some###24214texTTTT#S"
	tokenTTL   = 12 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId   int `json:"user_id"`
	UserRole int `json:"user_role"`
}

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (s *AuthService) CreateUser(user center.UserCreate) (int, error) {

	if user.UserType != int(center.PATIENT) {
		responceUserId, err := getUserIDFromToken(user.ResponceUser)

		if err != nil {
			return 0, err
		}

		isAdmin, err := s.repo.CheckUser(responceUserId)

		if err != nil || isAdmin != true {
			return 0, err
		}
	}

	user.Password = s.generatePasswordHash(user.Password)

	return s.repo.CreateUser(user)
}

func (s *AuthService) GenerateToken(username, password string) (string, error) {
	user, err := s.repo.GetUser(username, s.generatePasswordHash(password))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserId:   user.Id,
		UserRole: user.UserType,
	})

	return token.SignedString([]byte(signingKey))
}

func (s *AuthService) ParseToken(accessToken string) (int, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
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
		return 0, errors.New("token claims are not of type *tokenClaims")
	}

	return claims.UserId, nil
}

func (s *AuthService) generatePasswordHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func getUserIDFromToken(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("Failed to extract claims from the token")
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", errors.New("User ID not found in the token")
	}
	return userID, nil
}
