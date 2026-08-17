package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
	deptRepo *repository.DepartmentRepository
}

func NewUserService(userRepo *repository.UserRepository, deptRepo *repository.DepartmentRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		deptRepo: deptRepo,
	}
}

func (s *UserService) List(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, errors.New("id del usuario es requerido")
	}
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) Create(ctx context.Context, u *model.User, rawPassword string) error {
	if u.UserName == "" || u.UserCI == "" || u.UserNick == "" || rawPassword == "" {
		return errors.New("los campos user_name, user_ci, user_nick y contraseña son requeridos")
	}

	if u.DepartmentID != nil && *u.DepartmentID != "" {
		dept, err := s.deptRepo.GetByID(ctx, *u.DepartmentID)
		if err != nil {
			return fmt.Errorf("error validando departamento: %w", err)
		}
		if dept == nil {
			return errors.New("el departamento especificado no existe")
		}
	} else {
		u.DepartmentID = nil
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error encriptando contraseña: %w", err)
	}
	u.PasswordHash = string(hashedBytes)

	u.Active = true
	u.RequiresPasswordChange = true

	return s.userRepo.Create(ctx, u)
}

func (s *UserService) Update(ctx context.Context, u *model.User) error {
	if u.ID == "" || u.UserName == "" || u.UserCI == "" || u.UserNick == "" {
		return errors.New("los campos id, user_name, user_ci y user_nick son requeridos para actualizar")
	}

	existing, err := s.userRepo.GetByID(ctx, u.ID)
	if err != nil {
		return fmt.Errorf("error verificando usuario existente: %w", err)
	}
	if existing == nil {
		return errors.New("el usuario a actualizar no existe")
	}

	if u.DepartmentID != nil && *u.DepartmentID != "" {
		dept, err := s.deptRepo.GetByID(ctx, *u.DepartmentID)
		if err != nil {
			return fmt.Errorf("error validando departamento: %w", err)
		}
		if dept == nil {
			return errors.New("el departamento especificado no existe")
		}
	} else {
		u.DepartmentID = nil
	}

	return s.userRepo.Update(ctx, u)
}

func (s *UserService) Deactivate(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id del usuario es requerido")
	}

	existing, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error verificando usuario: %w", err)
	}
	if existing == nil {
		return errors.New("el usuario especificado no existe")
	}

	return s.userRepo.Deactivate(ctx, id)
}

func (s *UserService) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return s.userRepo.ListRoles(ctx)
}

func (s *UserService) ListPermissions(ctx context.Context) ([]*model.Permission, error) {
	return s.userRepo.ListPermissions(ctx)
}

func (s *UserService) UpdateRolePermissions(ctx context.Context, roleName string, permissionIDs []string) error {
	if roleName == "" {
		return errors.New("el nombre del rol es requerido")
	}
	return s.userRepo.UpdateRolePermissions(ctx, roleName, permissionIDs)
}
