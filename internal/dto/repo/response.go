package repo

import (
	"github.com/jamius19/postbranch/internal/db"
	"time"
)

type Response struct {
	ID        *int32        `json:"id"`
	Name      string        `json:"name"`
	PgVersion int32         `json:"pgVersion"`
	Status    db.RepoStatus `json:"status"`
	Output    *string       `json:"output"`
	Pool      Pool          `json:"pool"`
	Branches  []Branch      `json:"branches"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type Branch struct {
	ID        *int32            `json:"id"`
	Name      string            `json:"name"`
	Status    db.BranchStatus   `json:"status"`
	PgStatus  db.BranchPgStatus `json:"pgStatus"`
	Port      int32             `json:"port"`
	ParentID  *int32            `json:"parentId"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type Pool struct {
	ID        *int32 `json:"id"`
	Type      string `json:"type"`
	SizeInMb  int64  `json:"sizeInMb"`
	Path      string `json:"path"`
	MountPath string `json:"mountPath"`
}
