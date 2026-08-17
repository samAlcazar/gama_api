package model

import (
	"time"
)

type User struct {
	ID                     string     `json:"id"`
	UserName               string     `json:"user_name"`
	UserCI                 string     `json:"user_ci"`
	UserEmail              *string    `json:"user_email,omitempty"`
	UserPhone              *string    `json:"user_phone,omitempty"`
	DepartmentID           *string    `json:"department_id,omitempty"`
	Charge                 *string    `json:"charge,omitempty"`
	UserNick               string     `json:"user_nick"`
	PasswordHash           string     `json:"-"`
	UserPrincipalRole      string     `json:"user_principal_role"`
	Active                 bool       `json:"active"`
	RequiresPasswordChange bool       `json:"requires_password_change"`
	CreatedAt              time.Time  `json:"created_at"`
	LastAccess             *time.Time `json:"last_access,omitempty"`
	FailedAttempts         int        `json:"failed_attempts"`
	LockedUntil            *time.Time `json:"locked_until,omitempty"`
}

type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions,omitempty"`
}

type Permission struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

type UpdateRolePermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}

type UserWithPermissions struct {
	User        *User    `json:"user"`
	Permissions []string `json:"permissions"`
}
