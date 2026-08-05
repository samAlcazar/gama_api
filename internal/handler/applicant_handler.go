package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/service"
)

type ApplicantHandler struct {
	applicantService *service.ApplicantService
}

func NewApplicantHandler(applicantService *service.ApplicantService) *ApplicantHandler {
	return &ApplicantHandler{applicantService: applicantService}
}

func (h *ApplicantHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req model.CreateApplicantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	applicant, err := h.applicantService.Create(r.Context(), &req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, applicant)
}

func (h *ApplicantHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	applicants, err := h.applicantService.List(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, applicants)
}

func (h *ApplicantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id en la ruta")
		return
	}

	applicant, err := h.applicantService.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, applicant)
}
