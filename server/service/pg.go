package service

import (
	"context"
)

func PgCount(ctx context.Context) (int64, error) {
	pg, err := f.CountPg(ctx)

	if err != nil {
		log.Error(err)
		return -1, err
	}

	return pg, nil
}
