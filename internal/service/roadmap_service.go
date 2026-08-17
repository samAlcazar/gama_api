package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type RoadmapService struct {
	roadmapRepo   *repository.RoadmapRepository
	applicantRepo *repository.ApplicantRepository
	deptRepo      *repository.DepartmentRepository
	userRepo      *repository.UserRepository
}

func NewRoadmapService(
	roadmapRepo *repository.RoadmapRepository,
	applicantRepo *repository.ApplicantRepository,
	deptRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *RoadmapService {
	return &RoadmapService{
		roadmapRepo:   roadmapRepo,
		applicantRepo: applicantRepo,
		deptRepo:      deptRepo,
		userRepo:      userRepo,
	}
}

func (s *RoadmapService) Create(ctx context.Context, creatorUserID string, req *model.CreateRoadmapRequest) (*model.RoadmapDetail, error) {
	if strings.TrimSpace(req.Subject) == "" {
		return nil, errors.New("el asunto/resumen del trámite es obligatorio")
	}

	priority := strings.ToUpper(strings.TrimSpace(req.Priority))
	if priority == "" {
		priority = "MEDIA"
	}
	if priority != "ALTA" && priority != "MEDIA" && priority != "BAJA" {
		return nil, errors.New("la prioridad debe ser 'ALTA', 'MEDIA' o 'BAJA'")
	}

	pagesCount := req.PagesCount
	if pagesCount <= 0 {
		pagesCount = 1
	}

	if strings.TrimSpace(req.DestinationDeptID) == "" {
		return nil, errors.New("el departamento de destino inicial es obligatorio")
	}

	destDept, err := s.deptRepo.GetByID(ctx, req.DestinationDeptID)
	if err != nil {
		return nil, err
	}
	if destDept == nil {
		return nil, fmt.Errorf("el departamento de destino %s no existe", req.DestinationDeptID)
	}

	// Obtener datos del usuario creador para sacar su departamento de origen
	creatorUser, err := s.userRepo.GetByID(ctx, creatorUserID)
	if err != nil {
		return nil, fmt.Errorf("error verificando usuario creador: %w", err)
	}
	if creatorUser == nil {
		return nil, errors.New("usuario creador no encontrado")
	}

	// Gestión de Solicitante (si envió un nuevo solicitante o eligió uno existente)
	var applicantID *string
	if req.ApplicantID != nil && strings.TrimSpace(*req.ApplicantID) != "" {
		app, err := s.applicantRepo.GetByID(ctx, *req.ApplicantID)
		if err != nil {
			return nil, err
		}
		if app == nil {
			return nil, fmt.Errorf("el solicitante con id %s no existe", *req.ApplicantID)
		}
		applicantID = &app.ID
	} else if req.NewApplicant != nil {
		if strings.TrimSpace(req.NewApplicant.FullName) == "" || strings.TrimSpace(req.NewApplicant.CINIT) == "" {
			return nil, errors.New("nombre y CI/NIT son obligatorios para el nuevo solicitante")
		}
		// Buscar si ya existe por CI/NIT
		existingApp, err := s.applicantRepo.GetByCINit(ctx, strings.TrimSpace(req.NewApplicant.CINIT))
		if err != nil {
			return nil, err
		}
		if existingApp != nil {
			applicantID = &existingApp.ID
		} else {
			newApp := &model.Applicant{
				FullName: strings.TrimSpace(req.NewApplicant.FullName),
				CINIT:    strings.TrimSpace(req.NewApplicant.CINIT),
				Email:    req.NewApplicant.Email,
				Phone:    req.NewApplicant.Phone,
			}
			if err := s.applicantRepo.Create(ctx, newApp); err != nil {
				return nil, fmt.Errorf("error creando nuevo solicitante: %w", err)
			}
			applicantID = &newApp.ID
		}
	}

	currentYear := time.Now().Year()
	roadmapNumber, err := s.roadmapRepo.GetNextRoadmapNumber(ctx, currentYear)
	if err != nil {
		return nil, err
	}

	originDeptID := creatorUser.DepartmentID

	rm := &model.Roadmap{
		RoadmapNumber:      roadmapNumber,
		ManagementYear:     currentYear,
		ProcedureCode:      req.ProcedureCode,
		PagesCount:         pagesCount,
		OriginDepartmentID: originDeptID,
		Subject:            strings.TrimSpace(req.Subject),
		Priority:           priority,
		ApplicantID:        applicantID,
		Status:             "EN_RECORRIDO",
		CreatedBy:          creatorUserID,
	}

	firstStep := &model.RoadmapMovement{
		DestinationDepartmentID: req.DestinationDeptID,
		AssignedUserID:          req.AssignedUserID,
		Instruction:             req.Instruction,
		SignedBy:                &creatorUserID,
	}

	if err := s.roadmapRepo.Create(ctx, rm, firstStep); err != nil {
		return nil, err
	}

	return s.roadmapRepo.GetByID(ctx, rm.ID)
}

func (s *RoadmapService) ListVisible(ctx context.Context, userID string) ([]*model.Roadmap, error) {
	return s.roadmapRepo.ListVisible(ctx, userID)
}

func (s *RoadmapService) GetInbox(ctx context.Context, userID string) ([]*model.InboxItem, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("usuario no encontrado")
	}

	if user.UserPrincipalRole == "ADMIN" {
		return s.roadmapRepo.GetInbox(ctx, nil)
	}

	return s.roadmapRepo.GetInbox(ctx, user.DepartmentID)
}

func (s *RoadmapService) GetByID(ctx context.Context, id string) (*model.RoadmapDetail, error) {
	detail, err := s.roadmapRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, errors.New("hoja de ruta no encontrada")
	}
	return detail, nil
}

func (s *RoadmapService) Derive(ctx context.Context, roadmapID string, currentUserID string, req *model.DeriveRoadmapRequest) (*model.RoadmapMovement, error) {
	if strings.TrimSpace(req.DestinationDeptID) == "" {
		return nil, errors.New("el departamento de destino es obligatorio para derivar")
	}

	destDept, err := s.deptRepo.GetByID(ctx, req.DestinationDeptID)
	if err != nil {
		return nil, err
	}
	if destDept == nil {
		return nil, fmt.Errorf("el departamento de destino %s no existe", req.DestinationDeptID)
	}

	if req.AssignedUserID != nil && strings.TrimSpace(*req.AssignedUserID) != "" {
		assignedUser, err := s.userRepo.GetByID(ctx, *req.AssignedUserID)
		if err != nil {
			return nil, err
		}
		if assignedUser == nil {
			return nil, fmt.Errorf("el usuario asignado %s no existe", *req.AssignedUserID)
		}
	}

	detail, err := s.roadmapRepo.GetByID(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, errors.New("hoja de ruta no encontrada")
	}

	if detail.Roadmap.Status == "CONCLUIDO" || detail.Roadmap.Status == "ARCHIVADO" {
		return nil, fmt.Errorf("no se puede derivar una hoja de ruta con estado %s", detail.Roadmap.Status)
	}

	return s.roadmapRepo.Derive(ctx, roadmapID, currentUserID, req)
}

func (s *RoadmapService) UpdateStatus(ctx context.Context, roadmapID string, status string) error {
	st := strings.ToUpper(strings.TrimSpace(status))
	if st != "RESUELTO" && st != "CONCLUIDO" && st != "ARCHIVADO" && st != "RECHAZADO" {
		return errors.New("estado no válido. Debe ser 'RESUELTO', 'CONCLUIDO', 'ARCHIVADO' o 'RECHAZADO'")
	}
	return s.roadmapRepo.UpdateStatus(ctx, roadmapID, st)
}
