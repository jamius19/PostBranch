package dto

type RepoInit struct {
	Name     string `json:"name" validate:"required"`
	Path     string `json:"path" validate:"required"`
	RepoType string `json:"repoType" validate:"oneof=block virtual"`
	Size     int    `json:"size" validate:"required_if=RepoType virtual,gte=1"`
	SizeUnit string `json:"sizeUnit" validate:"required_if=RepoType virtual,oneof=KB MB GB"`
}
