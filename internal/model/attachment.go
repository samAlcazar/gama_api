package model

import "time"

type RoadmapAttachment struct {
	ID          string    `json:"id"`
	RoadmapID   string    `json:"roadmap_id"`
	MovementID  *string   `json:"movement_id,omitempty"`
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"-"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	PagesCount  int       `json:"pages_count"`
	Description *string   `json:"description,omitempty"`
	UploadedBy  string    `json:"uploaded_by"`
	Uploader    *User     `json:"uploader,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
