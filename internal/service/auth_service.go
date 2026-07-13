package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/samAlcazar/gama_api/internal/config"
	"github.com/samAlcazar/gama_api/internal/model"
	"github.com/samAlcazar/gama_api/internal/repository"
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepository
}

type CustomClaims struct {
	Nick        string   `json:"nick"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

func (s *AuthService) Login(ctx context.Context, nick, password string) (*model.UserWithPermissions, string, error) {
	u, err := s.userRepo.GetByNick(ctx, nick)
	if err != nil {
		return nil, "", fmt.Errorf("error al obtener usuario: %w", err)
	}

	if u == nil {
		return nil, "", errors.New("credenciales inválidas")
	}

	if !u.Active {
		return nil, "", errors.New("el usuario está inactivo")
	}

	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return nil, "", fmt.Errorf("cuenta bloqueada temporalmente hasta %s", u.LockedUntil.Format(time.RFC3339))
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", errors.New("credenciales inválidas")
	}

	perms, err := s.userRepo.GetPermissionsByRole(ctx, u.UserPrincipalRole)
	if err != nil {
		return nil, "", fmt.Errorf("error al obtener permisos: %w", err)
	}

	token, err := s.GenerateToken(u, perms)
	if err != nil {
		return nil, "", fmt.Errorf("error al generar token: %w", err)
	}

	if err := s.userRepo.UpdateLastAccess(ctx, u.ID); err != nil {
		fmt.Printf("advertencia: no se pudo actualizar last_access: %v\n", err)
	}

	return &model.UserWithPermissions{
		User:        u,
		Permissions: perms,
	}, token, nil
}

func (s *AuthService) GenerateToken(user *model.User, permissions []string) (string, error) {
	expirationTime := time.Now().Add(8 * time.Hour)

	claims := &CustomClaims{
		Nick:        user.UserNick,
		Role:        user.UserPrincipalRole,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenString, err := token.SignedString(s.cfg.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("error firmando token con llave privada: %w", err)
	}

	return tokenString, nil
}
