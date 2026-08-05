package model

import "time"

type Applicant struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	CINIT     string    `json:"ci_nit"`
	Email     *string   `json:"email,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateApplicantRequest struct {
	FullName string  `json:"full_name"`
	CINIT    string  `json:"ci_nit"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
}
