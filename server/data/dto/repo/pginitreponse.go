package repo

type PgInitResponseDto struct {
	ClusterSizeInMb int64 `json:"clusterSizeInMb" validate:"required,min=1"`
	PgInitDto
}
