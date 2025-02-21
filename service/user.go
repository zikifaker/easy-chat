package service

import (
	"context"
	"easy-chat/config"
	"easy-chat/dao"
	"easy-chat/entity"
	"easy-chat/request"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrInvalidPassword       = errors.New("invalid password")
	ErrFailedToGenerateToken = errors.New("failed to generate token")
)

const tokenExpirationHour = 24

func UserLogin(ctx context.Context, request *request.UserLoginRequest) (string, error) {
	user, err := dao.GetUserByUsername(request.Username)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	}

	user.LastLogin = time.Now()
	if err := dao.UpdateUser(user); err != nil {
		return "", err
	}

	token, err := generateToken(user)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFailedToGenerateToken, err)
	}

	return token, nil
}

func generateToken(user *entity.User) (string, error) {
	expirationTime := time.Now().Add(tokenExpirationHour * time.Hour)

	claims := &jwt.StandardClaims{
		Issuer:    user.Username,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	cfg := config.Get()
	tokenString, err := token.SignedString([]byte(cfg.SecretKey.JWT))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
