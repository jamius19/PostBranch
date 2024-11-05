package conversion

import "github.com/jamius19/postbranch/data/dao"

func SplitRepoRow(row *dao.GetRepoRow) (*dao.Repo, *dao.ZfsPool, *dao.Pg) {
	repo := dao.Repo{
		ID:        row.RepoID,
		Name:      row.RepoName,
		PoolID:    row.PoolID,
		CreatedAt: row.RepoCreatedAt,
		UpdatedAt: row.RepoUpdatedAt,
	}

	pool := dao.ZfsPool{
		ID:        row.PoolID,
		Path:      row.PoolPath,
		SizeInMb:  row.PoolSizeInMb,
		Name:      row.PoolName,
		MountPath: row.PoolMountPath,
		PoolType:  row.PoolType,
		CreatedAt: row.PoolCreatedAt,
		UpdatedAt: row.PoolUpdatedAt,
	}

	var pg *dao.Pg

	if row.PgID.Valid {
		pg = &dao.Pg{
			ID:             row.PgID.Int64,
			PgPath:         row.PgPath.String,
			Version:        row.PgVersion.Int64,
			StopPg:         row.PgStopPg.Bool,
			PgUser:         row.PgPgUser.String,
			ConnectionType: row.PgConnectionType.String,
			Host:           row.PgHost,
			Port:           row.PgPort,
			Username:       row.PgUsername,
			Password:       row.PgPassword,
			Status:         row.PgStatus.String,
			Output:         row.PgOutput,
			RepoID:         row.RepoID,
			CreatedAt:      row.PgCreatedAt.Time,
			UpdatedAt:      row.PgUpdatedAt.Time,
		}
	}

	return &repo, &pool, pg
}
