package repo

import (
	"time"
)

type Response struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	RepoType  string    `json:"repoType"`
	SizeInMb  int64     `json:"sizeInMb"`
	PoolID    int64     `json:"poolId"`
	Pg        *Pg       `json:"pg"`
	Branches  []Branch  `json:"branches"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Pg struct {
	PgID    int64   `json:"id"`
	Version int64   `json:"version"`
	Status  string  `json:"status"`
	Output  *string `json:"output"`
}

type Branch struct {
	Id        int64     `json:"id"`
	Name      string    `json:"name"`
	ParentId  *int64    `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
}
