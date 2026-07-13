package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samAlcazar/gama_api/internal/middleware"
	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/service"
)

type DepartmentHandler struct {
	deptService *service.DepartmentService
}

func NewDepartmentHandler(deptService *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	depts, err := h.deptService.List(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, depts)
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	role := middleware.GetUserRole(r.Context())
	if role != "ADMIN" {
		respondWithError(w, http.StatusForbidden, "acceso prohibido: requiere rol ADMIN")
		return
	}

	var d model.Department
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	if err := h.deptService.Create(r.Context(), &d); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, d)
}
