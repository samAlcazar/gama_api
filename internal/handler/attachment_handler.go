package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/samAlcazar/gama_api/internal/middleware"
	"github.com/samAlcazar/gama_api/internal/service"
)

type AttachmentHandler struct {
	attachmentService *service.AttachmentService
}

func NewAttachmentHandler(attachmentService *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachmentService: attachmentService}
}

func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	roadmapID := r.PathValue("id")
	if roadmapID == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id de la hoja de ruta en la ruta")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB Max
		respondWithError(w, http.StatusBadRequest, "error al procesar formulario multipart o archivo demasiado grande")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "se requiere el campo 'file' con el archivo a subir")
		return
	}
	defer file.Close()

	var movementID *string
	if mID := r.FormValue("movement_id"); mID != "" {
		movementID = &mID
	}

	var description *string
	if desc := r.FormValue("description"); desc != "" {
		description = &desc
	}

	pagesCount := 1
	if pc := r.FormValue("pages_count"); pc != "" {
		if parsedPC, err := strconv.Atoi(pc); err == nil && parsedPC > 0 {
			pagesCount = parsedPC
		}
	}

	att, err := h.attachmentService.Upload(r.Context(), roadmapID, userID, header, movementID, description, pagesCount)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, att)
}

func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	roadmapID := r.PathValue("id")
	if roadmapID == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id de la hoja de ruta en la ruta")
		return
	}

	attachments, err := h.attachmentService.ListByRoadmapID(r.Context(), roadmapID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, attachments)
}

func (h *AttachmentHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	attachmentID := r.PathValue("attachment_id")
	if attachmentID == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id del anexo en la ruta")
		return
	}

	att, err := h.attachmentService.GetByID(r.Context(), attachmentID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if _, err := os.Stat(att.FilePath); os.IsNotExist(err) {
		respondWithError(w, http.StatusNotFound, "el archivo físico no existe en el servidor")
		return
	}

	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+att.FileName+"\"")
	http.ServeFile(w, r, att.FilePath)
}

func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	attachmentID := r.PathValue("attachment_id")
	if attachmentID == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id del anexo en la ruta")
		return
	}

	if err := h.attachmentService.Delete(r.Context(), attachmentID); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "anexo eliminado correctamente"})
}
