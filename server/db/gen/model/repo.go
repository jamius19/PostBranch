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

type Repo struct {
	ID        *int32 `sql:"primary_key"`
	Name      string
	PoolID    int32
	CreatedAt time.Time
	UpdatedAt time.Time
}
