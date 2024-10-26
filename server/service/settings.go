package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jamius19/postbranch/data/fetch"
)

func GetSetting(ctx context.Context, key string) (*fetch.Setting, error) {
	settings, err := f.GetSetting(ctx, key)

	if err != nil && !errors.Is(sql.ErrNoRows, err) {
		log.Error(err)
		return nil, err
	} else if err != nil && errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	}

	return &settings, nil
}
