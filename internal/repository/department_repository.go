package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/model"
)

type DepartmentRepository struct {
	db *db.DB
}

func NewDepartmentRepository(db *db.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

func (r *DepartmentRepository) GetAll(ctx context.Context) ([]*model.Department, error) {
	query := `
		SELECT id, name, sigla, parent_department_id, level, active, created_at
		FROM departments
		ORDER BY level ASC, name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo todos los departamentos: %w", err)
	}
	defer rows.Close()

	var depts []*model.Department
	for rows.Next() {
		var d model.Department
		err := rows.Scan(&d.ID, &d.Name, &d.Sigla, &d.ParentDepartmentID, &d.Level, &d.Active, &d.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error escaneando departamento: %w", err)
		}
		depts = append(depts, &d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error en iteración de departamentos: %w", err)
	}

	return depts, nil
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id string) (*model.Department, error) {
	query := `
		SELECT id, name, sigla, parent_department_id, level, active, created_at
		FROM departments
		WHERE id = $1
	`
	var d model.Department
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.Name, &d.Sigla, &d.ParentDepartmentID, &d.Level, &d.Active, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo departamento por id: %w", err)
	}
	return &d, nil
}

func (r *DepartmentRepository) Create(ctx context.Context, d *model.Department) error {
	query := `
		INSERT INTO departments (name, sigla, parent_department_id, level, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.db.Pool.QueryRow(ctx, query, d.Name, d.Sigla, d.ParentDepartmentID, d.Level, d.Active).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creando departamento: %w", err)
	}
	return nil
}
