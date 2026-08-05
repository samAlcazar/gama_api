package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/samAlcazar/gama_api/internal/db"
	"github.com/samAlcazar/gama_api/internal/model"
)

type AttachmentRepository struct {
	db *db.DB
}

func NewAttachmentRepository(db *db.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(ctx context.Context, a *model.RoadmapAttachment) error {
	query := `
		INSERT INTO roadmap_attachments (
			roadmap_id, movement_id, file_name, file_path, file_size, mime_type, pages_count, description, uploaded_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	err := r.db.Pool.QueryRow(ctx, query,
		a.RoadmapID, a.MovementID, a.FileName, a.FilePath, a.FileSize, a.MimeType, a.PagesCount, a.Description, a.UploadedBy,
	).Scan(&a.ID, &a.CreatedAt)

	if err != nil {
		return fmt.Errorf("error registrando anexo en base de datos: %w", err)
	}
	return nil
}

func (r *AttachmentRepository) ListByRoadmapID(ctx context.Context, roadmapID string) ([]*model.RoadmapAttachment, error) {
	query := `
		SELECT 
			a.id, a.roadmap_id, a.movement_id, a.file_name, a.file_path, a.file_size, a.mime_type, a.pages_count, a.description, a.uploaded_by, a.created_at,
			u.user_name AS uploader_name
		FROM roadmap_attachments a
		LEFT JOIN users u ON a.uploaded_by = u.id
		WHERE a.roadmap_id = $1
		ORDER BY a.created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, roadmapID)
	if err != nil {
		return nil, fmt.Errorf("error listando anexos de la hoja de ruta: %w", err)
	}
	defer rows.Close()

	var list []*model.RoadmapAttachment
	for rows.Next() {
		var a model.RoadmapAttachment
		var uploaderName *string
		err := rows.Scan(
			&a.ID, &a.RoadmapID, &a.MovementID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.PagesCount, &a.Description, &a.UploadedBy, &a.CreatedAt,
			&uploaderName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando anexo: %w", err)
		}
		if uploaderName != nil {
			a.Uploader = &model.User{
				ID:       a.UploadedBy,
				UserName: *uploaderName,
			}
		}
		list = append(list, &a)
	}
	return list, nil
}

func (r *AttachmentRepository) GetByID(ctx context.Context, id string) (*model.RoadmapAttachment, error) {
	query := `
		SELECT 
			a.id, a.roadmap_id, a.movement_id, a.file_name, a.file_path, a.file_size, a.mime_type, a.pages_count, a.description, a.uploaded_by, a.created_at,
			u.user_name AS uploader_name
		FROM roadmap_attachments a
		LEFT JOIN users u ON a.uploaded_by = u.id
		WHERE a.id = $1
	`
	var a model.RoadmapAttachment
	var uploaderName *string
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.RoadmapID, &a.MovementID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.PagesCount, &a.Description, &a.UploadedBy, &a.CreatedAt,
		&uploaderName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo anexo por id: %w", err)
	}

	if uploaderName != nil {
		a.Uploader = &model.User{
			ID:       a.UploadedBy,
			UserName: *uploaderName,
		}
	}
	return &a, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM roadmap_attachments WHERE id = $1`
	cmd, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error eliminando registro de anexo: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("anexo no encontrado")
	}
	return nil
}
