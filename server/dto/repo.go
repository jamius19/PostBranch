package dto

import (
	"time"
)

type RepoInit struct {
	Name       string `json:"name" validate:"required"`
	BranchName string `json:"branchName" validate:"required"`
	Path       string `json:"path" validate:"required"`
	RepoType   string `json:"repoType" validate:"oneof=block virtual"`
	Size       int    `json:"size" validate:"required_if=RepoType virtual,gte=1"`
	SizeUnit   string `json:"sizeUnit" validate:"required_if=RepoType virtual,oneof=K M G"`
}

type RepoResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	PgID      *string   `json:"pg_id"`
	DatasetID int64     `json:"dataset_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
