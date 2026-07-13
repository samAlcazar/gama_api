package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type DepartmentService struct {
	deptRepo *repository.DepartmentRepository
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{deptRepo: deptRepo}
}

func (s *DepartmentService) List(ctx context.Context) ([]*model.Department, error) {
	return s.deptRepo.GetAll(ctx)
}

func (s *DepartmentService) GetByID(ctx context.Context, id string) (*model.Department, error) {
	if id == "" {
		return nil, errors.New("id del departamento es requerido")
	}
	return s.deptRepo.GetByID(ctx, id)
}

func (s *DepartmentService) Create(ctx context.Context, d *model.Department) error {
	if d.Name == "" {
		return errors.New("el nombre del departamento es requerido")
	}

	d.Level = 1
	d.Active = true

	if d.ParentDepartmentID != nil && *d.ParentDepartmentID != "" {
		parent, err := s.deptRepo.GetByID(ctx, *d.ParentDepartmentID)
		if err != nil {
			return fmt.Errorf("error validando departamento padre: %w", err)
		}
		if parent == nil {
			return errors.New("el departamento padre especificado no existe")
		}
		d.Level = parent.Level + 1
	} else {
		d.ParentDepartmentID = nil
	}

	return s.deptRepo.Create(ctx, d)
}
