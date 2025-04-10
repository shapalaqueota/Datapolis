package repository

import (
	"Datapolis/internal/models"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Открываем пул бд
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Устанавливаем роль по умолчанию если не указана
	if user.Role == "" {
		user.Role = models.RoleUser
	}

	err = r.db.QueryRow(ctx,
		`INSERT INTO users (username, password, email, role) 
		VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		user.Username, string(hashedPassword), user.Email, user.Role).Scan(&user.ID, &user.CreatedAt)
	return err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password, email, role, created_at 
		FROM users WHERE username = $1`, username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password, email, role, created_at 
		FROM users WHERE email = $1`, email).
		Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
