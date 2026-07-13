package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/model"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByNick(ctx context.Context, nick string) (*model.User, error) {
	query := `
		SELECT 
			id, user_name, user_ci, user_email, user_phone, department_id, 
			charge, user_nick, password_hash, user_principal_role, active, 
			requires_password_change, created_at, last_access, failed_attempts, locked_until
		FROM users
		WHERE user_nick = $1
	`
	var u model.User
	err := r.db.Pool.QueryRow(ctx, query, nick).Scan(
		&u.ID, &u.UserName, &u.UserCI, &u.UserEmail, &u.UserPhone, &u.DepartmentID,
		&u.Charge, &u.UserNick, &u.PasswordHash, &u.UserPrincipalRole, &u.Active,
		&u.RequiresPasswordChange, &u.CreatedAt, &u.LastAccess, &u.FailedAttempts, &u.LockedUntil,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo usuario por nick: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetPermissionsByRole(ctx context.Context, roleName string) ([]string, error) {
	query := `
		SELECT p.code
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_name = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, roleName)
	if err != nil {
		return nil, fmt.Errorf("error consultando permisos de rol: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("error escaneando código de permiso: %w", err)
		}
		permissions = append(permissions, code)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando sobre filas de permisos: %w", err)
	}

	return permissions, nil
}

func (r *UserRepository) UpdateLastAccess(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_access = NOW(), failed_attempts = 0, locked_until = NULL
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error actualizando último acceso: %w", err)
	}
	return nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	query := `
		SELECT 
			id, user_name, user_ci, user_email, user_phone, department_id, 
			charge, user_nick, password_hash, user_principal_role, active, 
			requires_password_change, created_at, last_access, failed_attempts, locked_until
		FROM users
		ORDER BY user_name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo todos los usuarios: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(
			&u.ID, &u.UserName, &u.UserCI, &u.UserEmail, &u.UserPhone, &u.DepartmentID,
			&u.Charge, &u.UserNick, &u.PasswordHash, &u.UserPrincipalRole, &u.Active,
			&u.RequiresPasswordChange, &u.CreatedAt, &u.LastAccess, &u.FailedAttempts, &u.LockedUntil,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando usuario: %w", err)
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterando sobre usuarios: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT 
			id, user_name, user_ci, user_email, user_phone, department_id, 
			charge, user_nick, password_hash, user_principal_role, active, 
			requires_password_change, created_at, last_access, failed_attempts, locked_until
		FROM users
		WHERE id = $1
	`
	var u model.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.UserName, &u.UserCI, &u.UserEmail, &u.UserPhone, &u.DepartmentID,
		&u.Charge, &u.UserNick, &u.PasswordHash, &u.UserPrincipalRole, &u.Active,
		&u.RequiresPasswordChange, &u.CreatedAt, &u.LastAccess, &u.FailedAttempts, &u.LockedUntil,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo usuario por id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	query := `
		INSERT INTO users (
			user_name, user_ci, user_email, user_phone, department_id, 
			charge, user_nick, password_hash, user_principal_role, active, requires_password_change
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`
	err := r.db.Pool.QueryRow(
		ctx, query,
		u.UserName, u.UserCI, u.UserEmail, u.UserPhone, u.DepartmentID,
		u.Charge, u.UserNick, u.PasswordHash, u.UserPrincipalRole, u.Active, u.RequiresPasswordChange,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creando usuario en base de datos: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	query := `
		UPDATE users
		SET 
			user_name = $1, 
			user_ci = $2, 
			user_email = $3, 
			user_phone = $4, 
			department_id = $5, 
			charge = $6, 
			user_nick = $7, 
			user_principal_role = $8, 
			active = $9, 
			requires_password_change = $10
		WHERE id = $11
	`
	_, err := r.db.Pool.Exec(
		ctx, query,
		u.UserName, u.UserCI, u.UserEmail, u.UserPhone, u.DepartmentID,
		u.Charge, u.UserNick, u.UserPrincipalRole, u.Active, u.RequiresPasswordChange,
		u.ID,
	)
	if err != nil {
		return fmt.Errorf("error actualizando usuario: %w", err)
	}
	return nil
}

func (r *UserRepository) Deactivate(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET active = false
		WHERE id = $1
	`
	_, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error desactivando usuario: %w", err)
	}
	return nil
}
