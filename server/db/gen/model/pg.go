//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Pg struct {
	ID        *int32 `sql:"primary_key"`
	PgPath    string
	Version   int32
	Status    string
	Output    *string
	RepoID    int32
	CreatedAt time.Time
	UpdatedAt time.Time
}