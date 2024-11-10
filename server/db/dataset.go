package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
)

func CreateDataset(ctx context.Context, dataset model.ZfsDataset) (model.ZfsDataset, error) {
	var newDataset model.ZfsDataset

	stmt := table.ZfsDataset.INSERT(table.ZfsDataset.AllColumns).
		MODEL(dataset).
		RETURNING(table.ZfsDataset.AllColumns)

	log.Debugf("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &newDataset)
	if err != nil {
		log.Errorf("Can't insert dataset: %s", err)
		return model.ZfsDataset{}, err
	}

	return newDataset, nil
}

func GetDatasetByName(ctx context.Context, datasetName string) (model.ZfsDataset, error) {
	var dataset model.ZfsDataset

	stmt := table.ZfsDataset.SELECT(table.ZfsDataset.AllColumns).
		FROM(table.ZfsDataset).
		WHERE(table.ZfsDataset.Name.EQ(sqlite.String(datasetName)))

	log.Debugf("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &dataset)
	if err != nil {
		log.Warnf("Can't get dataset: %s", err)
		return model.ZfsDataset{}, err
	}

	return dataset, nil
}
