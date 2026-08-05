package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type ApplicantService struct {
	applicantRepo *repository.ApplicantRepository
}

func NewApplicantService(applicantRepo *repository.ApplicantRepository) *ApplicantService {
	return &ApplicantService{applicantRepo: applicantRepo}
}

func (s *ApplicantService) Create(ctx context.Context, req *model.CreateApplicantRequest) (*model.Applicant, error) {
	if strings.TrimSpace(req.FullName) == "" {
		return nil, errors.New("el nombre o razón social del solicitante es obligatorio")
	}
	if strings.TrimSpace(req.CINIT) == "" {
		return nil, errors.New("el C.I. o NIT del solicitante es obligatorio")
	}

	existing, err := s.applicantRepo.GetByCINit(ctx, strings.TrimSpace(req.CINIT))
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("ya existe un solicitante registrado con el C.I./NIT %s", req.CINIT)
	}

	applicant := &model.Applicant{
		FullName: strings.TrimSpace(req.FullName),
		CINIT:    strings.TrimSpace(req.CINIT),
		Email:    req.Email,
		Phone:    req.Phone,
	}

	if err := s.applicantRepo.Create(ctx, applicant); err != nil {
		return nil, err
	}

	return applicant, nil
}

func (s *ApplicantService) GetByID(ctx context.Context, id string) (*model.Applicant, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("el id del solicitante es obligatorio")
	}
	applicant, err := s.applicantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if applicant == nil {
		return nil, errors.New("solicitante no encontrado")
	}
	return applicant, nil
}

func (s *ApplicantService) List(ctx context.Context) ([]*model.Applicant, error) {
	return s.applicantRepo.List(ctx)
}
