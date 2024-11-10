package repo

import (
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
	ID      *int32  `json:"id"`
	Version int32   `json:"version"`
	Status  string  `json:"status"`
	Output  *string `json:"output"`
}

type Branch struct {
	ID        *int32    `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int32    `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
}

type Pool struct {
	ID       *int32 `json:"id"`
	Type     string `json:"type"`
	SizeInMb int64  `json:"sizeInMb"`
	Path     string `json:"path"`
}
