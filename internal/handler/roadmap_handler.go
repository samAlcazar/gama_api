package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samAlcazar/gama_api/internal/middleware"
	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/service"
)

type RoadmapHandler struct {
	roadmapService *service.RoadmapService
}

func NewRoadmapHandler(roadmapService *service.RoadmapService) *RoadmapHandler {
	return &RoadmapHandler{roadmapService: roadmapService}
}

func (h *RoadmapHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	var req model.CreateRoadmapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	detail, err := h.roadmapService.Create(r.Context(), userID, &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, detail)
}

func (h *RoadmapHandler) ListVisible(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	roadmaps, err := h.roadmapService.ListVisible(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, roadmaps)
}

func (h *RoadmapHandler) GetInbox(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	inbox, err := h.roadmapService.GetInbox(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, inbox)
}

func (h *RoadmapHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id de la hoja de ruta en la ruta")
		return
	}

	detail, err := h.roadmapService.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, detail)
}

func (h *RoadmapHandler) Derive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id de la hoja de ruta en la ruta")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	var req model.DeriveRoadmapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	newMovement, err := h.roadmapService.Derive(r.Context(), id, userID, &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, newMovement)
}

func (h *RoadmapHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id de la hoja de ruta en la ruta")
		return
	}

	var req model.UpdateRoadmapStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	if err := h.roadmapService.UpdateStatus(r.Context(), id, req.Status); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "estado actualizado correctamente"})
}
