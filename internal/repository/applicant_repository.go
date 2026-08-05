package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/model"
)

type ApplicantRepository struct {
	db *db.DB
}

func NewApplicantRepository(db *db.DB) *ApplicantRepository {
	return &ApplicantRepository{db: db}
}

func (r *ApplicantRepository) Create(ctx context.Context, a *model.Applicant) error {
	query := `
		INSERT INTO applicants (full_name, ci_nit, email, phone)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := r.db.Pool.QueryRow(ctx, query, a.FullName, a.CINIT, a.Email, a.Phone).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creando solicitante: %w", err)
	}
	return nil
}

func (r *ApplicantRepository) GetByID(ctx context.Context, id string) (*model.Applicant, error) {
	query := `
		SELECT id, full_name, ci_nit, email, phone, created_at
		FROM applicants
		WHERE id = $1
	`
	var a model.Applicant
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.FullName, &a.CINIT, &a.Email, &a.Phone, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo solicitante por id: %w", err)
	}
	return &a, nil
}

func (r *ApplicantRepository) GetByCINit(ctx context.Context, ciNit string) (*model.Applicant, error) {
	query := `
		SELECT id, full_name, ci_nit, email, phone, created_at
		FROM applicants
		WHERE ci_nit = $1
	`
	var a model.Applicant
	err := r.db.Pool.QueryRow(ctx, query, ciNit).Scan(
		&a.ID, &a.FullName, &a.CINIT, &a.Email, &a.Phone, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo solicitante por ci_nit: %w", err)
	}
	return &a, nil
}

func (r *ApplicantRepository) List(ctx context.Context) ([]*model.Applicant, error) {
	query := `
		SELECT id, full_name, ci_nit, email, phone, created_at
		FROM applicants
		ORDER BY full_name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listando solicitantes: %w", err)
	}
	defer rows.Close()

	var list []*model.Applicant
	for rows.Next() {
		var a model.Applicant
		if err := rows.Scan(&a.ID, &a.FullName, &a.CINIT, &a.Email, &a.Phone, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("error escaneando solicitante: %w", err)
		}
		list = append(list, &a)
	}
	return list, nil
}
