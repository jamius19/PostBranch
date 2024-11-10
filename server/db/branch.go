package db

import (
	"context"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
)

func CreateBranch(ctx context.Context, branch model.Branch) (model.Branch, error) {
	var newBranch model.Branch
	stmt := table.Branch.INSERT(table.Branch.AllColumns).
		MODEL(branch).
		RETURNING(table.Branch.AllColumns)

	log.Debugf("Query: %s", stmt.DebugSql())
	
	err := stmt.QueryContext(ctx, Db, &newBranch)
	if err != nil {
		log.Errorf("Can't create branch: %s", err)
		return model.Branch{}, err
	}

	return newBranch, nil
}
