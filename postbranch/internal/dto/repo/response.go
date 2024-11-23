package repo

import (
	db2 "github.com/jamius19/postbranch/internal/db"
	"time"
)

type Response struct {
	ID        *int32    `json:"id"`
	Name      string    `json:"name"`
	Pg        Pg        `json:"pg"`
	Pool      Pool      `json:"pool"`
	Branches  []Branch  `json:"branches"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Pg struct {
	ID      *int32       `json:"id"`
	Version int32        `json:"version"`
	Status  db2.PgStatus `json:"status"`
	Output  *string      `json:"output"`
}

type Branch struct {
	ID        *int32             `json:"id"`
	Name      string             `json:"name"`
	Status    db2.BranchStatus   `json:"status"`
	PgStatus  db2.BranchPgStatus `json:"pgStatus"`
	Port      int32              `json:"port"`
	ParentID  *int32             `json:"parentId"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type Pool struct {
	ID        *int32 `json:"id"`
	Type      string `json:"type"`
	SizeInMb  int64  `json:"sizeInMb"`
	Path      string `json:"path"`
	MountPath string `json:"mountPath"`
}