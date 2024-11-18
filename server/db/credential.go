package db

import (
	"context"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/jamius19/postbranch/db/gen/model"
	"github.com/jamius19/postbranch/db/gen/table"
)

func AddCredential(ctx context.Context, repoId int32, password string) (model.Credential, error) {
	var credential model.Credential

	stmt := table.Credential.
		INSERT(table.Credential.RepoID, table.Credential.Password, table.Credential.CreatedAt, table.Credential.UpdatedAt).
		VALUES(repoId, password, sqlite.CURRENT_TIME(), sqlite.CURRENT_TIME()).
		RETURNING(table.Credential.AllColumns)

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &credential)
	if err != nil {
		log.Errorf("Can't insert credential: %s", err)
		return model.Credential{}, err
	}

	return credential, nil
}

func GetCredentialByRepoId(ctx context.Context, repoId int32) (model.Credential, error) {
	var credential model.Credential

	stmt := table.Credential.
		SELECT(table.Credential.AllColumns).
		WHERE(table.Credential.RepoID.EQ(sqlite.Int32(repoId)))

	log.Tracef("Query: %s", stmt.DebugSql())

	err := stmt.QueryContext(ctx, Db, &credential)
	if err != nil {
		log.Errorf("Can't get credential for repo %d: %s", repoId, err)
		return model.Credential{}, err
	}

	return credential, nil
}
