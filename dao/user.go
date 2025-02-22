package dao

import (
	"easy-chat/entity"
	"easy-chat/request"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
)

func CreateUser(request *request.UserRegisterRequest) (*entity.User, error) {
	if err := checkUserExists(request); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username: request.Username,
		Email:    request.Email,
		Password: string(passwordHash),
	}
	result := DB.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

func GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	if result := DB.Where("username = ?", username).First(&user); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func UpdateUser(user *entity.User) error {
	if err := DB.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func checkUserExists(request *request.UserRegisterRequest) error {
	var user entity.User

	if result := DB.Where("username = ?", request.Username).First(&user); result.Error == nil {
		return fmt.Errorf("%w: %s", ErrUsernameAlreadyExists, request.Username)
	}

	if result := DB.Where("email = ?", request.Email).First(&user); result.Error == nil {
		return fmt.Errorf("%w: %s", ErrEmailAlreadyExists, request.Email)
	}

	return nil
}
