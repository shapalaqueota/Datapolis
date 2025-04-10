package service

import (
	"Datapolis/internal/models"
	"Datapolis/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists      = errors.New("пользователь с таким именем или email уже существует")
	ErrInvalidLogin    = errors.New("неверное имя пользователя или пароль")
	ErrInvalidUserData = errors.New("неверные данные пользователя")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, user *models.User) error {
	existingUser, err := s.repo.GetByUsername(ctx, user.Username)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	existingUser, err = s.repo.GetByEmail(ctx, user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	return s.repo.Create(ctx, user)
}

func (s *UserService) Login(ctx context.Context, username, password string) (*models.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidLogin
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidLogin
	}

	user.Password = ""
	return user, nil
}
