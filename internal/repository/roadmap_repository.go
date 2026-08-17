package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/model"
)

type RoadmapRepository struct {
	db *db.DB
}

func NewRoadmapRepository(db *db.DB) *RoadmapRepository {
	return &RoadmapRepository{db: db}
}

func (r *RoadmapRepository) GetNextRoadmapNumber(ctx context.Context, managementYear int) (string, error) {
	query := `
		SELECT COALESCE(
			MAX(CAST(SPLIT_PART(SPLIT_PART(roadmap_number, '/', 1), '-', 2) AS INTEGER)), 0
		) + 1
		FROM roadmaps
		WHERE management_year = $1
	`
	var nextSeq int
	err := r.db.Pool.QueryRow(ctx, query, managementYear).Scan(&nextSeq)
	if err != nil {
		return "", fmt.Errorf("error generando correlativo de hoja de ruta: %w", err)
	}
	return fmt.Sprintf("HR-%04d/%d", nextSeq, managementYear), nil
}

func (r *RoadmapRepository) Create(ctx context.Context, rm *model.Roadmap, firstStep *model.RoadmapMovement) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error iniciando transacción para crear hoja de ruta: %w", err)
	}
	defer tx.Rollback(ctx)

	queryRoadmap := `
		INSERT INTO roadmaps (
			roadmap_number, management_year, procedure_code, pages_count,
			origin_department_id, subject, priority, applicant_id, status, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	err = tx.QueryRow(ctx, queryRoadmap,
		rm.RoadmapNumber, rm.ManagementYear, rm.ProcedureCode, rm.PagesCount,
		rm.OriginDepartmentID, rm.Subject, rm.Priority, rm.ApplicantID, rm.Status, rm.CreatedBy,
	).Scan(&rm.ID, &rm.CreatedAt)

	if err != nil {
		return fmt.Errorf("error insertando hoja de ruta: %w", err)
	}

	queryStep := `
		INSERT INTO roadmap_movements (
			roadmap_id, step_number, destination_department_id, assigned_user_id,
			entry_at, instruction, signed_by, status
		)
		VALUES ($1, 1, $2, $3, NOW(), $4, $5, 'PENDIENTE')
		RETURNING id, entry_at, created_at
	`
	firstStep.RoadmapID = rm.ID
	firstStep.StepNumber = 1
	firstStep.Status = "PENDIENTE"

	err = tx.QueryRow(ctx, queryStep,
		rm.ID, firstStep.DestinationDepartmentID, firstStep.AssignedUserID,
		firstStep.Instruction, firstStep.SignedBy,
	).Scan(&firstStep.ID, &firstStep.EntryAt, &firstStep.CreatedAt)

	if err != nil {
		return fmt.Errorf("error insertando primer movimiento de hoja de ruta: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error confirmando transacción de hoja de ruta: %w", err)
	}

	return nil
}

func (r *RoadmapRepository) ListVisible(ctx context.Context, userID string) ([]*model.Roadmap, error) {
	query := `
		SELECT 
			r.id, r.roadmap_number, r.management_year, r.procedure_code, r.pages_count,
			r.origin_department_id, r.subject, r.priority, r.applicant_id, r.status, r.created_by, r.created_at,
			d.name AS origin_dept_name, d.sigla AS origin_dept_sigla,
			app.full_name AS applicant_name, app.ci_nit AS applicant_ci_nit,
			u.user_name AS creator_name
		FROM fn_get_visible_roadmaps($1) r
		LEFT JOIN departments d ON r.origin_department_id = d.id
		LEFT JOIN applicants app ON r.applicant_id = app.id
		LEFT JOIN users u ON r.created_by = u.id
		ORDER BY r.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error listando hojas de ruta visibles: %w", err)
	}
	defer rows.Close()

	var list []*model.Roadmap
	for rows.Next() {
		var rm model.Roadmap
		var deptName, deptSigla, appName, appCINIT, creatorName *string

		err := rows.Scan(
			&rm.ID, &rm.RoadmapNumber, &rm.ManagementYear, &rm.ProcedureCode, &rm.PagesCount,
			&rm.OriginDepartmentID, &rm.Subject, &rm.Priority, &rm.ApplicantID, &rm.Status, &rm.CreatedBy, &rm.CreatedAt,
			&deptName, &deptSigla, &appName, &appCINIT, &creatorName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando hoja de ruta: %w", err)
		}

		if rm.OriginDepartmentID != nil && deptName != nil {
			rm.OriginDepartment = &model.Department{
				ID:    *rm.OriginDepartmentID,
				Name:  *deptName,
				Sigla: deptSigla,
			}
		}
		if rm.ApplicantID != nil && appName != nil {
			rm.Applicant = &model.Applicant{
				ID:       *rm.ApplicantID,
				FullName: *appName,
				CINIT:    *appCINIT,
			}
		}
		if creatorName != nil {
			rm.CreatedByUser = &model.User{
				ID:       rm.CreatedBy,
				UserName: *creatorName,
			}
		}

		list = append(list, &rm)
	}
	return list, nil
}

func (r *RoadmapRepository) GetInbox(ctx context.Context, departmentID *string) ([]*model.InboxItem, error) {
	query := `
		SELECT 
			movement_id, roadmap_id, roadmap_number, management_year, procedure_code,
			subject, priority, pages_count, roadmap_status, applicant_name, applicant_ci_nit,
			step_number, destination_department_id, destination_department_name,
			assigned_user_id, assigned_user_name, entry_at, instruction, movement_status
		FROM vw_user_inbox
		WHERE ($1::UUID IS NULL OR destination_department_id = $1)
		ORDER BY entry_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, departmentID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo bandeja de entrada: %w", err)
	}
	defer rows.Close()

	var inbox []*model.InboxItem
	for rows.Next() {
		var item model.InboxItem
		err := rows.Scan(
			&item.MovementID, &item.RoadmapID, &item.RoadmapNumber, &item.ManagementYear, &item.ProcedureCode,
			&item.Subject, &item.Priority, &item.PagesCount, &item.RoadmapStatus, &item.ApplicantName, &item.ApplicantCINIT,
			&item.StepNumber, &item.DestinationDepartmentID, &item.DestinationDepartmentName,
			&item.AssignedUserID, &item.AssignedUserName, &item.EntryAt, &item.Instruction, &item.MovementStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando item de bandeja de entrada: %w", err)
		}
		inbox = append(inbox, &item)
	}
	return inbox, nil
}

func (r *RoadmapRepository) GetByID(ctx context.Context, id string) (*model.RoadmapDetail, error) {
	queryRoadmap := `
		SELECT 
			r.id, r.roadmap_number, r.management_year, r.procedure_code, r.pages_count,
			r.origin_department_id, r.subject, r.priority, r.applicant_id, r.status, r.created_by, r.created_at,
			d.name AS origin_dept_name, d.sigla AS origin_dept_sigla,
			app.full_name AS applicant_name, app.ci_nit AS applicant_ci_nit, app.email AS applicant_email, app.phone AS applicant_phone,
			u.user_name AS creator_name
		FROM roadmaps r
		LEFT JOIN departments d ON r.origin_department_id = d.id
		LEFT JOIN applicants app ON r.applicant_id = app.id
		LEFT JOIN users u ON r.created_by = u.id
		WHERE r.id = $1
	`
	var rm model.Roadmap
	var deptName, deptSigla, appName, appCINIT, appEmail, appPhone, creatorName *string

	err := r.db.Pool.QueryRow(ctx, queryRoadmap, id).Scan(
		&rm.ID, &rm.RoadmapNumber, &rm.ManagementYear, &rm.ProcedureCode, &rm.PagesCount,
		&rm.OriginDepartmentID, &rm.Subject, &rm.Priority, &rm.ApplicantID, &rm.Status, &rm.CreatedBy, &rm.CreatedAt,
		&deptName, &deptSigla, &appName, &appCINIT, &appEmail, &appPhone, &creatorName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo detalle de hoja de ruta: %w", err)
	}

	if rm.OriginDepartmentID != nil && deptName != nil {
		rm.OriginDepartment = &model.Department{
			ID:    *rm.OriginDepartmentID,
			Name:  *deptName,
			Sigla: deptSigla,
		}
	}
	if rm.ApplicantID != nil && appName != nil {
		rm.Applicant = &model.Applicant{
			ID:       *rm.ApplicantID,
			FullName: *appName,
			CINIT:    *appCINIT,
			Email:    appEmail,
			Phone:    appPhone,
		}
	}
	if creatorName != nil {
		rm.CreatedByUser = &model.User{
			ID:       rm.CreatedBy,
			UserName: *creatorName,
		}
	}

	movements, err := r.GetMovementsByRoadmapID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &model.RoadmapDetail{
		Roadmap:   &rm,
		Movements: movements,
	}, nil
}

func (r *RoadmapRepository) GetMovementsByRoadmapID(ctx context.Context, roadmapID string) ([]*model.RoadmapMovement, error) {
	query := `
		SELECT 
			rm.id, rm.roadmap_id, rm.step_number, rm.destination_department_id, rm.assigned_user_id,
			rm.entry_at, rm.exit_at, rm.instruction, rm.signed_by, rm.status, rm.created_at,
			d.name AS dest_dept_name, d.sigla AS dest_dept_sigla,
			u_assigned.user_name AS assigned_user_name,
			u_signed.user_name AS signed_user_name
		FROM roadmap_movements rm
		JOIN departments d ON rm.destination_department_id = d.id
		LEFT JOIN users u_assigned ON rm.assigned_user_id = u_assigned.id
		LEFT JOIN users u_signed ON rm.signed_by = u_signed.id
		WHERE rm.roadmap_id = $1
		ORDER BY rm.step_number ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo movimientos de la hoja de ruta: %w", err)
	}
	defer rows.Close()

	var list []*model.RoadmapMovement
	for rows.Next() {
		var m model.RoadmapMovement
		var destDeptName, destDeptSigla, assignedName, signedName *string

		err := rows.Scan(
			&m.ID, &m.RoadmapID, &m.StepNumber, &m.DestinationDepartmentID, &m.AssignedUserID,
			&m.EntryAt, &m.ExitAt, &m.Instruction, &m.SignedBy, &m.Status, &m.CreatedAt,
			&destDeptName, &destDeptSigla, &assignedName, &signedName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando movimiento de hoja de ruta: %w", err)
		}

		if destDeptName != nil {
			m.DestinationDepartment = &model.Department{
				ID:    m.DestinationDepartmentID,
				Name:  *destDeptName,
				Sigla: destDeptSigla,
			}
		}
		if m.AssignedUserID != nil && assignedName != nil {
			m.AssignedUser = &model.User{
				ID:       *m.AssignedUserID,
				UserName: *assignedName,
			}
		}
		if m.SignedBy != nil && signedName != nil {
			m.SignedByUser = &model.User{
				ID:       *m.SignedBy,
				UserName: *signedName,
			}
		}

		list = append(list, &m)
	}
	return list, nil
}

func (r *RoadmapRepository) Derive(ctx context.Context, roadmapID string, currentUserID string, req *model.DeriveRoadmapRequest) (*model.RoadmapMovement, error) {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error iniciando transacción de derivación: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Obtener el último paso de la hoja de ruta
	var lastStepID string
	var lastStepNumber int
	queryLastStep := `
		SELECT id, step_number
		FROM roadmap_movements
		WHERE roadmap_id = $1
		ORDER BY step_number DESC
		LIMIT 1
	`
	err = tx.QueryRow(ctx, queryLastStep, roadmapID).Scan(&lastStepID, &lastStepNumber)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo el último paso del expediente: %w", err)
	}

	// 2. Cerrar el paso anterior (exit_at = NOW(), signed_by = currentUserID, status = 'DERIVADO')
	now := time.Now()
	queryCloseStep := `
		UPDATE roadmap_movements
		SET exit_at = $1, signed_by = $2, status = 'DERIVADO'
		WHERE id = $3
	`
	_, err = tx.Exec(ctx, queryCloseStep, now, currentUserID, lastStepID)
	if err != nil {
		return nil, fmt.Errorf("error cerrando paso anterior de la hoja de ruta: %w", err)
	}

	// 3. Crear el nuevo paso
	newStepNumber := lastStepNumber + 1
	var newMovement model.RoadmapMovement
	queryNewStep := `
		INSERT INTO roadmap_movements (
			roadmap_id, step_number, destination_department_id, assigned_user_id,
			entry_at, instruction, status
		)
		VALUES ($1, $2, $3, $4, NOW(), $5, 'PENDIENTE')
		RETURNING id, roadmap_id, step_number, destination_department_id, assigned_user_id, entry_at, instruction, status, created_at
	`
	err = tx.QueryRow(ctx, queryNewStep,
		roadmapID, newStepNumber, req.DestinationDeptID, req.AssignedUserID, req.Instruction,
	).Scan(
		&newMovement.ID, &newMovement.RoadmapID, &newMovement.StepNumber, &newMovement.DestinationDepartmentID,
		&newMovement.AssignedUserID, &newMovement.EntryAt, &newMovement.Instruction, &newMovement.Status, &newMovement.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creando nuevo paso de derivación: %w", err)
	}

	// 4. Actualizar estado de la hoja de ruta a EN_RECORRIDO
	queryUpdateRoadmap := `
		UPDATE roadmaps
		SET status = 'EN_RECORRIDO'
		WHERE id = $1
	`
	_, err = tx.Exec(ctx, queryUpdateRoadmap, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("error actualizando estado global de la hoja de ruta: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error confirmando transacción de derivación: %w", err)
	}

	return &newMovement, nil
}

func (r *RoadmapRepository) UpdateStatus(ctx context.Context, roadmapID string, status string) error {
	query := `
		UPDATE roadmaps
		SET status = $1
		WHERE id = $2
	`
	cmd, err := r.db.Pool.Exec(ctx, query, status, roadmapID)
	if err != nil {
		return fmt.Errorf("error actualizando estado de la hoja de ruta: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("hoja de ruta no encontrada")
	}
	return nil
}
