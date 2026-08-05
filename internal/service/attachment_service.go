package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type AttachmentService struct {
	attachmentRepo *repository.AttachmentRepository
	roadmapRepo    *repository.RoadmapRepository
	uploadDir      string
}

func NewAttachmentService(
	attachmentRepo *repository.AttachmentRepository,
	roadmapRepo *repository.RoadmapRepository,
	uploadDir string,
) *AttachmentService {
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	return &AttachmentService{
		attachmentRepo: attachmentRepo,
		roadmapRepo:    roadmapRepo,
		uploadDir:      uploadDir,
	}
}

func (s *AttachmentService) Upload(
	ctx context.Context,
	roadmapID string,
	uploaderID string,
	fileHeader *multipart.FileHeader,
	movementID *string,
	description *string,
	pagesCount int,
) (*model.RoadmapAttachment, error) {
	roadmapDetail, err := s.roadmapRepo.GetByID(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	if roadmapDetail == nil {
		return nil, errors.New("hoja de ruta no encontrada")
	}

	if fileHeader.Size > 50*1024*1024 { // 50MB Max
		return nil, errors.New("el archivo excede el tamaño máximo permitido de 50MB")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("error abriendo archivo adjunto: %w", err)
	}
	defer file.Close()

	// Crear subdirectorio por año y mes
	now := time.Now()
	subDir := filepath.Join(s.uploadDir, now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(subDir, 0755); err != nil {
		return nil, fmt.Errorf("error creando directorio de almacenamiento: %w", err)
	}

	ext := filepath.Ext(fileHeader.Filename)
	uniqueFileName := fmt.Sprintf("%d%s", now.UnixNano(), ext)
	dstPath := filepath.Join(subDir, uniqueFileName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("error creando archivo en disco: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("error guardando contenido del archivo: %w", err)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if pagesCount <= 0 {
		pagesCount = 1
	}

	attachment := &model.RoadmapAttachment{
		RoadmapID:   roadmapID,
		MovementID:  movementID,
		FileName:    fileHeader.Filename,
		FilePath:    dstPath,
		FileSize:    fileHeader.Size,
		MimeType:    mimeType,
		PagesCount:  pagesCount,
		Description: description,
		UploadedBy:  uploaderID,
	}

	if err := s.attachmentRepo.Create(ctx, attachment); err != nil {
		_ = os.Remove(dstPath) // Limpiar archivo en caso de fallo DB
		return nil, err
	}

	return attachment, nil
}

func (s *AttachmentService) ListByRoadmapID(ctx context.Context, roadmapID string) ([]*model.RoadmapAttachment, error) {
	return s.attachmentRepo.ListByRoadmapID(ctx, roadmapID)
}

func (s *AttachmentService) GetByID(ctx context.Context, id string) (*model.RoadmapAttachment, error) {
	att, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, errors.New("anexo no encontrado")
	}
	return att, nil
}

func (s *AttachmentService) Delete(ctx context.Context, id string) error {
	att, err := s.attachmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if att == nil {
		return errors.New("anexo no encontrado")
	}

	if err := s.attachmentRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = os.Remove(att.FilePath) // Borrar archivo del disco
	return nil
}
