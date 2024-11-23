package pg

type ValidationResponseDto[T HostImportReqDto] struct {
	ClusterSizeInMb int64 `json:"clusterSizeInMb" validate:"required,min=1"`
	PgConfig        T     `json:"pgConfig"`
}
