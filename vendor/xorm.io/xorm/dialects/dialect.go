// Copyright 2019 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dialects

import (
	"context"
	"fmt"
	"strings"
	"time"

	"xorm.io/xorm/core"
	"xorm.io/xorm/schemas"
)

// URI represents an uri to visit database
type URI struct {
	DBType  schemas.DBType
	Proto   string
	Host    string
	Port    string
	DBName  string
	User    string
	Passwd  string
	Charset string
	Laddr   string
	Raddr   string
	Timeout time.Duration
	Schema  string
}

// SetSchema set schema
func (uri *URI) SetSchema(schema string) {
	// hack me
	if uri.DBType == schemas.POSTGRES {
		uri.Schema = strings.TrimSpace(schema)
	}
}

// enumerates all autoincr mode
const (
	IncrAutoincrMode = iota
	SequenceAutoincrMode
)

// DialectFeatures represents a dialect parameters
type DialectFeatures struct {
	AutoincrMode int // 0 autoincrement column, 1 sequence
}

// Dialect represents a kind of database
type Dialect interface {
	Init(*URI) error
	URI() *URI
	Version(ctx context.Context, queryer core.Queryer) (*schemas.Version, error)
	Features() *DialectFeatures

	SQLType(*schemas.Column) string
	Alias(string) string       // return what a sql type's alias of
	ColumnTypeKind(string) int // database column type kind

	IsReserved(string) bool
	Quoter() schemas.Quoter
	SetQuotePolicy(quotePolicy QuotePolicy)

	AutoIncrStr() string

	GetIndexes(queryer core.Queryer, ctx context.Context, tableName string) (map[string]*schemas.Index, error)
	IndexCheckSQL(tableName, idxName string) (string, []interface{})
	CreateIndexSQL(tableName string, index *schemas.Index) string
	DropIndexSQL(tableName string, index *schemas.Index) string

	GetTables(queryer core.Queryer, ctx context.Context) ([]*schemas.Table, error)
	IsTableExist(queryer core.Queryer, ctx context.Context, tableName string) (bool, error)
	CreateTableSQL(ctx context.Context, queryer core.Queryer, table *schemas.Table, tableName string) (string, bool, error)
	DropTableSQL(tableName string) (string, bool)

	CreateSequenceSQL(ctx context.Context, queryer core.Queryer, seqName string) (string, error)
	IsSequenceExist(ctx context.Context, queryer core.Queryer, seqName string) (bool, error)
	DropSequenceSQL(seqName string) (string, error)

	GetColumns(queryer core.Queryer, ctx context.Context, tableName string) ([]string, map[string]*schemas.Column, error)
	IsColumnExist(queryer core.Queryer, ctx context.Context, tableName string, colName string) (bool, error)
	AddColumnSQL(tableName string, col *schemas.Column) string
	ModifyColumnSQL(tableName string, col *schemas.Column) string

	ForUpdateSQL(query string) string

	Filters() []Filter
	SetParams(params map[string]string)
}

// Base represents a basic dialect and all real dialects could embed this struct
type Base struct {
	dialect Dialect
	uri     *URI
	quoter  schemas.Quoter
}

// Alias returned col itself
func (db *Base) Alias(col string) string {
	return col
}

// Quoter returns the current database Quoter
func (db *Base) Quoter() schemas.Quoter {
	return db.quoter
}

// Init initialize the dialect
func (db *Base) Init(dialect Dialect, uri *URI) error {
	db.dialect, db.uri = dialect, uri
	return nil
}

// URI returns the uri of database
func (db *Base) URI() *URI {
	return db.uri
}

// CreateTableSQL implements Dialect
func (db *Base) CreateTableSQL(ctx context.Context, queryer core.Queryer, table *schemas.Table, tableName string) (string, bool, error) {
	if tableName == "" {
		tableName = table.Name
	}

	quoter := db.dialect.Quoter()
	var b strings.Builder
	b.WriteString("CREATE TABLE IF NOT EXISTS ")
	if err := quoter.QuoteTo(&b, tableName); err != nil {
		return "", false, err
	}
	b.WriteString(" (")

	for i, colName := range table.ColumnsSeq() {
		col := table.GetColumn(colName)
		s, _ := ColumnString(db.dialect, col, col.IsPrimaryKey && len(table.PrimaryKeys) == 1)
		b.WriteString(s)

		if i != len(table.ColumnsSeq())-1 {
			b.WriteString(", ")
		}
	}

	if len(table.PrimaryKeys) > 1 {
		b.WriteString(", PRIMARY KEY (")
		b.WriteString(quoter.Join(table.PrimaryKeys, ","))
		b.WriteString(")")
	}

	b.WriteString(")")

	return b.String(), false, nil
}

func (db *Base) CreateSequenceSQL(ctx context.Context, queryer core.Queryer, seqName string) (string, error) {
	return fmt.Sprintf(`CREATE SEQUENCE %s 
	minvalue 1
	   nomaxvalue
	   start with 1
	   increment by 1
	   nocycle
	nocache`, seqName), nil
}

func (db *Base) IsSequenceExist(ctx context.Context, queryer core.Queryer, seqName string) (bool, error) {
	return false, fmt.Errorf("unsupported sequence feature")
}

func (db *Base) DropSequenceSQL(seqName string) (string, error) {
	return fmt.Sprintf("DROP SEQUENCE %s", seqName), nil
}

// DropTableSQL returns drop table SQL
func (db *Base) DropTableSQL(tableName string) (string, bool) {
	quote := db.dialect.Quoter().Quote
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", quote(tableName)), true
}

// HasRecords returns true if the SQL has records returned
func (db *Base) HasRecords(queryer core.Queryer, ctx context.Context, query string, args ...interface{}) (bool, error) {
	rows, err := queryer.QueryContext(ctx, query, args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	return false, rows.Err()
}

// IsColumnExist returns true if the column of the table exist
func (db *Base) IsColumnExist(queryer core.Queryer, ctx context.Context, tableName, colName string) (bool, error) {
	quote := db.dialect.Quoter().Quote
	query := fmt.Sprintf(
		"SELECT %v FROM %v.%v WHERE %v = ? AND %v = ? AND %v = ?",
		quote("COLUMN_NAME"),
		quote("INFORMATION_SCHEMA"),
		quote("COLUMNS"),
		quote("TABLE_SCHEMA"),
		quote("TABLE_NAME"),
		quote("COLUMN_NAME"),
	)
	return db.HasRecords(queryer, ctx, query, db.uri.DBName, tableName, colName)
}

// AddColumnSQL returns a SQL to add a column
func (db *Base) AddColumnSQL(tableName string, col *schemas.Column) string {
	s, _ := ColumnString(db.dialect, col, true)
	return fmt.Sprintf("ALTER TABLE %s ADD %s", db.dialect.Quoter().Quote(tableName), s)
}

// CreateIndexSQL returns a SQL to create index
func (db *Base) CreateIndexSQL(tableName string, index *schemas.Index) string {
	quoter := db.dialect.Quoter()
	var unique string
	var idxName string
	if index.Type == schemas.UniqueType {
		unique = " UNIQUE"
	}
	idxName = index.XName(tableName)
	return fmt.Sprintf("CREATE%s INDEX %v ON %v (%v)", unique,
		quoter.Quote(idxName), quoter.Quote(tableName),
		quoter.Join(index.Cols, ","))
}

// DropIndexSQL returns a SQL to drop index
func (db *Base) DropIndexSQL(tableName string, index *schemas.Index) string {
	quote := db.dialect.Quoter().Quote
	var name string
	if index.IsRegular {
		name = index.XName(tableName)
	} else {
		name = index.Name
	}
	return fmt.Sprintf("DROP INDEX %v ON %s", quote(name), quote(tableName))
}

// ModifyColumnSQL returns a SQL to modify SQL
func (db *Base) ModifyColumnSQL(tableName string, col *schemas.Column) string {
	s, _ := ColumnString(db.dialect, col, false)
	return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", db.quoter.Quote(tableName), s)
}

// ForUpdateSQL returns for updateSQL
func (db *Base) ForUpdateSQL(query string) string {
	return query + " FOR UPDATE"
}

// SetParams set params
func (db *Base) SetParams(params map[string]string) {
}

var (
	dialects = map[string]func() Dialect{}
)

// RegisterDialect register database dialect
func RegisterDialect(dbName schemas.DBType, dialectFunc func() Dialect) {
	if dialectFunc == nil {
		panic("core: Register dialect is nil")
	}
	dialects[strings.ToLower(string(dbName))] = dialectFunc // !nashtsai! allow override dialect
}

// QueryDialect query if registered database dialect
func QueryDialect(dbName schemas.DBType) Dialect {
	if d, ok := dialects[strings.ToLower(string(dbName))]; ok {
		return d()
	}
	return nil
}

func regDrvsNDialects() bool {
	providedDrvsNDialects := map[string]struct {
		dbType     schemas.DBType
		getDriver  func() Driver
		getDialect func() Dialect
	}{
		"mssql":    {"mssql", func() Driver { return &odbcDriver{} }, func() Dialect { return &mssql{} }},
		"odbc":     {"mssql", func() Driver { return &odbcDriver{} }, func() Dialect { return &mssql{} }}, // !nashtsai! TODO change this when supporting MS Access
		"mysql":    {"mysql", func() Driver { return &mysqlDriver{} }, func() Dialect { return &mysql{} }},
		"mymysql":  {"mysql", func() Driver { return &mymysqlDriver{} }, func() Dialect { return &mysql{} }},
		"postgres": {"postgres", func() Driver { return &pqDriver{} }, func() Dialect { return &postgres{} }},
		"pgx":      {"postgres", func() Driver { return &pqDriverPgx{} }, func() Dialect { return &postgres{} }},
		"sqlite3":  {"sqlite3", func() Driver { return &sqlite3Driver{} }, func() Dialect { return &sqlite3{} }},
		"sqlite":   {"sqlite3", func() Driver { return &sqlite3Driver{} }, func() Dialect { return &sqlite3{} }},
		"oci8":     {"oracle", func() Driver { return &oci8Driver{} }, func() Dialect { return &oracle{} }},
		"godror":   {"oracle", func() Driver { return &godrorDriver{} }, func() Dialect { return &oracle{} }},
	}

	for driverName, v := range providedDrvsNDialects {
		if driver := QueryDriver(driverName); driver == nil {
			RegisterDriver(driverName, v.getDriver())
			RegisterDialect(v.dbType, v.getDialect)
		}
	}
	return true
}

func init() {
	regDrvsNDialects()
}

// ColumnString generate column description string according dialect
func ColumnString(dialect Dialect, col *schemas.Column, includePrimaryKey bool) (string, error) {
	bd := strings.Builder{}

	if err := dialect.Quoter().QuoteTo(&bd, col.Name); err != nil {
		return "", err
	}

	if err := bd.WriteByte(' '); err != nil {
		return "", err
	}

	if _, err := bd.WriteString(dialect.SQLType(col)); err != nil {
		return "", err
	}

	if includePrimaryKey && col.IsPrimaryKey {
		if _, err := bd.WriteString(" PRIMARY KEY"); err != nil {
			return "", err
		}
		if col.IsAutoIncrement {
			if err := bd.WriteByte(' '); err != nil {
				return "", err
			}
			if _, err := bd.WriteString(dialect.AutoIncrStr()); err != nil {
				return "", err
			}
		}
	}

	if !col.DefaultIsEmpty {
		if _, err := bd.WriteString(" DEFAULT "); err != nil {
			return "", err
		}
		if col.Default == "" {
			if _, err := bd.WriteString("''"); err != nil {
				return "", err
			}
		} else {
			if _, err := bd.WriteString(col.Default); err != nil {
				return "", err
			}
		}
	}

	if col.Nullable {
		if _, err := bd.WriteString(" NULL"); err != nil {
			return "", err
		}
	} else {
		if _, err := bd.WriteString(" NOT NULL"); err != nil {
			return "", err
		}
	}

	return bd.String(), nil
}
