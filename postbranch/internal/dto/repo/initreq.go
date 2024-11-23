package repo

import (
	"github.com/jamius19/postbranch/internal/dto/pg"
)

const MinSizeInMb = 300

type Config struct {
	Name     string `json:"name" validate:"required,min=1,excludesall= "`
	Path     string `json:"path" validate:"required,min=1,excludesall= "`
	RepoType string `json:"repoType" validate:"oneof=block virtual,excludesall= "`
	SizeInMb int64  `json:"sizeInMb" validate:"required_if=RepoType virtual"`
}

type InitDto[T pg.HostImportReqDto] struct {
	RepoConfig Config `json:"repoConfig"`
	PgConfig   T      `json:"pgConfig"`
}

func (initDto *InitDto[T]) GetName() string {
	return initDto.RepoConfig.Name
}

func (initDto *InitDto[T]) GetPath() string {
	return initDto.RepoConfig.Path
}

func (initDto *InitDto[T]) GetRepoType() string {
	return initDto.RepoConfig.RepoType
}

func (initDto *InitDto[T]) GetSizeInMb() int64 {
	return initDto.RepoConfig.SizeInMb
}

// Info needs to be replaced with the Config type TODO
type Info interface {
	GetName() string
	GetPath() string
	GetRepoType() string
	GetSizeInMb() int64
}
