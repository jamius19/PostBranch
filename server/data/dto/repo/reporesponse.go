package repo

import (
	"time"
)

type RepoResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	RepoType  string    `json:"repoType"`
	SizeInMb  int64     `json:"sizeInMb"`
	PgID      *int64    `json:"pgId"`
	PoolID    int64     `json:"poolId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
