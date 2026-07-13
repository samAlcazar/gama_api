package model

import "time"

type Department struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	Sigla              *string    `json:"sigla,omitempty"`
	ParentDepartmentID *string    `json:"parent_department_id,omitempty"`
	Level              int        `json:"level"`
	Active             bool       `json:"active"`
	CreatedAt          time.Time  `json:"created_at"`
}
