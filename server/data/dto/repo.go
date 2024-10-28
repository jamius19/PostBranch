package dto

import (
	"time"
)

type RepoInit struct {
	Name     string `json:"name" validate:"required"`
	Path     string `json:"path" validate:"required"`
	RepoType string `json:"repoType" validate:"oneof=block virtual"`
	Size     int64  `json:"size" validate:"required_if=RepoType virtual,gte=1"`
	SizeUnit string `json:"sizeUnit" validate:"required_if=RepoType virtual,oneof=K M G"`
}

type RepoResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	RepoType  string    `json:"repoType"`
	Size      int64     `json:"size"`
	SizeUnit  string    `json:"sizeUnit"`
	PgID      *int64    `json:"pg_id"`
	PoolID    int64     `json:"pool_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
