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

var (
	ErrUserNotFound         = errors.New("пользователь не найден")
	ErrNoPermission         = errors.New("недостаточно прав для выполнения операции")
	ErrCannotDeactivateSelf = errors.New("невозможно деактивировать свой собственный аккаунт")
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
		refreshExpiresIn = 0
	}

	return &auth.TokenPair{
		AccessToken:      accessToken,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("пользователь не найден")
	}
	return user, nil
}

// getAllUsers получает всех пользователей
func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return nil, errors.New("пользователи не найдены")
	}
	return users, nil
}

// ---- Update user ----------------------------------------

func (s *UserService) UpdateUser(ctx context.Context, updaterID int, userToUpdate *models.User) error {
	updater, err := s.repo.GetByID(ctx, updaterID)
	if err != nil {
		return err
	}
	if updater == nil {
		return ErrUserNotFound
	}

	existingUser, err := s.repo.GetByID(ctx, userToUpdate.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	isAdmin := updater.Role == models.RoleAdmin
	isSelf := updaterID == userToUpdate.ID

	if !isSelf && !isAdmin {
		return ErrNoPermission
	}

	if isSelf && !isAdmin && updater.Role != userToUpdate.Role {
		userToUpdate.Role = updater.Role
	}

	if isSelf && !userToUpdate.IsActive {
		return ErrCannotDeactivateSelf
	}

	if existingUser.Username != userToUpdate.Username {
		userByUsername, err := s.repo.GetByUsername(ctx, userToUpdate.Username)
		if err != nil {
			return err
		}
		if userByUsername != nil {
			return ErrUserExists
		}
	}

	if existingUser.Email != userToUpdate.Email {
		userByEmail, err := s.repo.GetByEmail(ctx, userToUpdate.Email)
		if err != nil {
			return err
		}
		if userByEmail != nil {
			return ErrUserExists
		}
	}

	return s.repo.Update(ctx, userToUpdate)
}

func (s *UserService) UpdatePassword(ctx context.Context, updaterID int, userID int, newPassword string) error {
	updater, err := s.repo.GetByID(ctx, updaterID)
	if err != nil {
		return err
	}
	if updater == nil {
		return ErrUserNotFound
	}

	userToUpdate, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if userToUpdate == nil {
		return ErrUserNotFound
	}

	isAdmin := updater.Role == models.RoleAdmin
	isSelf := updaterID == userID

	if !isSelf && !isAdmin {
		return ErrNoPermission
	}

	if len(newPassword) < 6 {
		return errors.New("пароль должен содержать не менее 6 символов")
	}

	return s.repo.UpdatePassword(ctx, userID, newPassword)
}
