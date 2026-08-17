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

func (s *DepartmentService) Update(ctx context.Context, d *model.Department) error {
	if d.ID == "" || d.Name == "" {
		return errors.New("el id y el nombre del departamento son requeridos")
	}

	existing, err := s.deptRepo.GetByID(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("error obteniendo departamento: %w", err)
	}
	if existing == nil {
		return errors.New("el departamento especificado no existe")
	}

	d.Level = 1
	d.Active = existing.Active

	if d.ParentDepartmentID != nil && *d.ParentDepartmentID != "" {
		if *d.ParentDepartmentID == d.ID {
			return errors.New("un departamento no puede ser su propio departamento padre")
		}

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

	return s.deptRepo.Update(ctx, d)
}

func (s *DepartmentService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("el id del departamento es requerido")
	}

	existing, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error obteniendo departamento: %w", err)
	}
	if existing == nil {
		return errors.New("el departamento especificado no existe")
	}

	return s.deptRepo.Delete(ctx, id)
}
