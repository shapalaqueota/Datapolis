package service

import (
	"Datapolis/internal/auth"
	"Datapolis/internal/models"
	"Datapolis/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrUserExists      = errors.New("пользователь с таким именем или email уже существует")
	ErrInvalidUserData = errors.New("неверные данные пользователя")
)

type AuthService struct {
	userRepo *repository.UserRepository
}
type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
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

func (s *AuthService) Login(ctx context.Context, username, password string) (*auth.TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("пользователь не найден")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	tokenPair, err := auth.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	claims, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("недействительный refresh токен")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("пользователь не найден")
	}

	accessToken, expiresIn, err := auth.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshExpiresIn := int64(time.Until(claims.ExpiresAt.Time).Seconds())
	if refreshExpiresIn < 0 {
		refreshExpiresIn = 0 // На случай, если токен почти истек
	}

	return &auth.TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}
