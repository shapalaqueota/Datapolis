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

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, username, password, email, role, created_at 
		FROM users WHERE id = $1`, id).
		Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// get all users
func (r *UserRepository) GetAll(ctx context.Context) ([]*models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, username, password, email, role, created_at 
		FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

//-- Update user ----------------------------------------

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users 
         SET username = $1, email = $2, role = $3, is_active = $4, updated_at = NOW()
         WHERE id = $5`,
		user.Username, user.Email, user.Role, user.IsActive, user.ID)
	return err
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx,
		`UPDATE users 
         SET password = $1, updated_at = NOW()
         WHERE id = $2`,
		string(hashedPassword), userID)
	return err
}
