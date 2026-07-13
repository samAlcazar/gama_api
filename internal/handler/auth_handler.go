package handler

import (
	"encoding/json"
	"net/http"

	"github.com/samAlcazar/gama_api/internal/middleware"
	"github.com/samAlcazar/gama_api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

type loginRequest struct {
	UserNick string `json:"user_nick"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "cuerpo de solicitud inválido")
		return
	}

	if req.UserNick == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "se requiere user_nick y password")
		return
	}

	userWithPerms, token, err := h.authService.Login(r.Context(), req.UserNick, req.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, loginResponse{
		Token: token,
		User:  userWithPerms,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	userID := middleware.GetUserID(r.Context())
	nick := middleware.GetUserNick(r.Context())
	role := middleware.GetUserRole(r.Context())
	perms := middleware.GetUserPermissions(r.Context())

	response := map[string]interface{}{
		"id":          userID,
		"user_nick":   nick,
		"role":        role,
		"permissions": perms,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
