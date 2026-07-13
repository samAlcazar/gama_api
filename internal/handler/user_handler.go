package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type createUserRequest struct {
	model.User
	Password string `json:"password"`
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	users, err := h.userService.List(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id en la ruta")
		return
	}

	u, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if u == nil {
		respondWithError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	if err := h.userService.Create(r.Context(), &req.User, req.Password); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, req.User)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id en la ruta")
		return
	}

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	u.ID = id

	if err := h.userService.Update(r.Context(), &u); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (h *UserHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere el id en la ruta")
		return
	}

	if err := h.userService.Deactivate(r.Context(), id); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "usuario desactivado correctamente"})
}
