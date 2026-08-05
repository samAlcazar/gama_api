package model

import "time"

type Roadmap struct {
	ID                 string      `json:"id"`
	RoadmapNumber      string      `json:"roadmap_number"`
	ManagementYear     int         `json:"management_year"`
	ProcedureCode      *string     `json:"procedure_code,omitempty"`
	PagesCount         int         `json:"pages_count"`
	OriginDepartmentID *string     `json:"origin_department_id,omitempty"`
	OriginDepartment   *Department `json:"origin_department,omitempty"`
	Subject            string      `json:"subject"`
	Priority           string      `json:"priority"`
	ApplicantID        *string     `json:"applicant_id,omitempty"`
	Applicant          *Applicant  `json:"applicant,omitempty"`
	Status             string      `json:"status"`
	CreatedBy          string      `json:"created_by"`
	CreatedByUser      *User       `json:"created_by_user,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
}

type RoadmapMovement struct {
	ID                      string      `json:"id"`
	RoadmapID               string      `json:"roadmap_id"`
	StepNumber              int         `json:"step_number"`
	DestinationDepartmentID string      `json:"destination_department_id"`
	DestinationDepartment   *Department `json:"destination_department,omitempty"`
	AssignedUserID          *string     `json:"assigned_user_id,omitempty"`
	AssignedUser            *User       `json:"assigned_user,omitempty"`
	EntryAt                 time.Time   `json:"entry_at"`
	ExitAt                  *time.Time  `json:"exit_at,omitempty"`
	Instruction             *string     `json:"instruction,omitempty"`
	SignedBy                *string     `json:"signed_by,omitempty"`
	SignedByUser            *User       `json:"signed_by_user,omitempty"`
	Status                  string      `json:"status"`
	CreatedAt               time.Time   `json:"created_at"`
}

type RoadmapDetail struct {
	Roadmap   *Roadmap           `json:"roadmap"`
	Movements []*RoadmapMovement `json:"movements"`
}

type CreateRoadmapRequest struct {
	ProcedureCode     *string                 `json:"procedure_code,omitempty"`
	PagesCount        int                     `json:"pages_count"`
	Subject           string                  `json:"subject"`
	Priority          string                  `json:"priority"` // 'ALTA', 'MEDIA', 'BAJA'
	ApplicantID       *string                 `json:"applicant_id,omitempty"`
	NewApplicant      *CreateApplicantRequest `json:"new_applicant,omitempty"`
	DestinationDeptID string                  `json:"destination_department_id"`
	AssignedUserID    *string                 `json:"assigned_user_id,omitempty"`
	Instruction       *string                 `json:"instruction,omitempty"`
}

type DeriveRoadmapRequest struct {
	DestinationDeptID string  `json:"destination_department_id"`
	AssignedUserID    *string `json:"assigned_user_id,omitempty"`
	Instruction       *string `json:"instruction,omitempty"`
}

type UpdateRoadmapStatusRequest struct {
	Status string `json:"status"` // 'RESUELTO', 'CONCLUIDO', 'ARCHIVADO', 'RECHAZADO'
}

type InboxItem struct {
	MovementID                string    `json:"movement_id"`
	RoadmapID                 string    `json:"roadmap_id"`
	RoadmapNumber             string    `json:"roadmap_number"`
	ManagementYear            int       `json:"management_year"`
	ProcedureCode             *string   `json:"procedure_code,omitempty"`
	Subject                   string    `json:"subject"`
	Priority                  string    `json:"priority"`
	PagesCount                int       `json:"pages_count"`
	RoadmapStatus             string    `json:"roadmap_status"`
	ApplicantName             *string   `json:"applicant_name,omitempty"`
	ApplicantCINIT            *string   `json:"applicant_ci_nit,omitempty"`
	StepNumber                int       `json:"step_number"`
	DestinationDepartmentID   string    `json:"destination_department_id"`
	DestinationDepartmentName string    `json:"destination_department_name"`
	AssignedUserID            *string   `json:"assigned_user_id,omitempty"`
	AssignedUserName          *string   `json:"assigned_user_name,omitempty"`
	EntryAt                   time.Time `json:"entry_at"`
	Instruction               *string   `json:"instruction,omitempty"`
	MovementStatus            string    `json:"movement_status"`
}
