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

type ZfsPool struct {
	ID        *int32 `sql:"primary_key"`
	Path      string
	SizeInMb  int64
	Name      string
	MountPath string
	PoolType  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
