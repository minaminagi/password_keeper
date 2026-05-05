package config

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/services/sqlite"
)

//go:embed migrations/001_init.sql
var tableInitSQL string

//go:embed migrations/002_index.sql
var indexInitSQL string

//go:embed migrations/003_recovery.sql
var recoveryInitSQL string

var db *sqlite.SQLiteService

func initDB(dbSource string) error {
	db = sqlite.NewWithConfig(&sqlite.Config{DBSource: dbSource})
	if err := db.Open(); err != nil {
		return err
	}
	if err := execSQLStatements(tableInitSQL); err != nil {
		return fmt.Errorf("init tables: %w", err)
	}
	if err := execSQLStatements(indexInitSQL); err != nil {
		return fmt.Errorf("init indexes: %w", err)
	}
	if err := execSQLStatements(recoveryInitSQL); err != nil {
		return fmt.Errorf("init recovery tables: %w", err)
	}
	return nil
}

func execSQLStatements(sqlText string) error {
	sqls := strings.SplitSeq(sqlText, ";")
	for sql := range sqls {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if err := db.Execute(sql); err != nil {
			return fmt.Errorf("execute sql failed: %w", err)
		}
	}
	return nil
}

func SQLiteService(dbSource string) (*sqlite.SQLiteService, error) {
	err := initDB(dbSource)
	return db, err
}
