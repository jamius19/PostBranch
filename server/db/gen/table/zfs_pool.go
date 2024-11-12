//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/sqlite"
)

var ZfsPool = newZfsPoolTable("", "zfs_pool", "")

type zfsPoolTable struct {
	sqlite.Table

	// Columns
	ID        sqlite.ColumnInteger
	Path      sqlite.ColumnString
	SizeInMb  sqlite.ColumnInteger
	Name      sqlite.ColumnString
	MountPath sqlite.ColumnString
	PoolType  sqlite.ColumnString
	CreatedAt sqlite.ColumnTimestamp
	UpdatedAt sqlite.ColumnTimestamp

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
}

type ZfsPoolTable struct {
	zfsPoolTable

	EXCLUDED zfsPoolTable
}

// AS creates new ZfsPoolTable with assigned alias
func (a ZfsPoolTable) AS(alias string) *ZfsPoolTable {
	return newZfsPoolTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ZfsPoolTable with assigned schema name
func (a ZfsPoolTable) FromSchema(schemaName string) *ZfsPoolTable {
	return newZfsPoolTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ZfsPoolTable with assigned table prefix
func (a ZfsPoolTable) WithPrefix(prefix string) *ZfsPoolTable {
	return newZfsPoolTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ZfsPoolTable with assigned table suffix
func (a ZfsPoolTable) WithSuffix(suffix string) *ZfsPoolTable {
	return newZfsPoolTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newZfsPoolTable(schemaName, tableName, alias string) *ZfsPoolTable {
	return &ZfsPoolTable{
		zfsPoolTable: newZfsPoolTableImpl(schemaName, tableName, alias),
		EXCLUDED:     newZfsPoolTableImpl("", "excluded", ""),
	}
}

func newZfsPoolTableImpl(schemaName, tableName, alias string) zfsPoolTable {
	var (
		IDColumn        = sqlite.IntegerColumn("id")
		PathColumn      = sqlite.StringColumn("path")
		SizeInMbColumn  = sqlite.IntegerColumn("size_in_mb")
		NameColumn      = sqlite.StringColumn("name")
		MountPathColumn = sqlite.StringColumn("mount_path")
		PoolTypeColumn  = sqlite.StringColumn("pool_type")
		CreatedAtColumn = sqlite.TimestampColumn("created_at")
		UpdatedAtColumn = sqlite.TimestampColumn("updated_at")
		allColumns      = sqlite.ColumnList{IDColumn, PathColumn, SizeInMbColumn, NameColumn, MountPathColumn, PoolTypeColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns  = sqlite.ColumnList{PathColumn, SizeInMbColumn, NameColumn, MountPathColumn, PoolTypeColumn, CreatedAtColumn, UpdatedAtColumn}
	)

	return zfsPoolTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:        IDColumn,
		Path:      PathColumn,
		SizeInMb:  SizeInMbColumn,
		Name:      NameColumn,
		MountPath: MountPathColumn,
		PoolType:  PoolTypeColumn,
		CreatedAt: CreatedAtColumn,
		UpdatedAt: UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}