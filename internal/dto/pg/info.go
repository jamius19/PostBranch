package pg

import "github.com/jamius19/postbranch/internal/db"

type Info interface {
	GetAdapter() db.RepoPgAdapter
	GetPgPath() string
	GetVersion() int32
}
