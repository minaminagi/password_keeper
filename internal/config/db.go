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

var db *sqlite.SQLiteService

func initDB(dbSource string) error {
	db = sqlite.NewWithConfig(&sqlite.Config{DBSource: dbSource})
	if err := db.Open(); err != nil {
		return err
	}
	err := initTables()
	err = initIndex()
	return err
}

func initTables() error {
	sqls := strings.SplitSeq(tableInitSQL, ";")
	for sql := range sqls {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if err := db.Execute(sql); err != nil {
			return fmt.Errorf("init table failed: %w", err)
		}
	}
	return nil
}

func initIndex() error {
	sqls := strings.SplitSeq(indexInitSQL, ";")
	for sql := range sqls {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if err := db.Execute(sql); err != nil {
			return fmt.Errorf("init table failed: %w", err)
		}
	}
	return nil
}

func SQLiteService(dbSource string) (*sqlite.SQLiteService, error) {
	err := initDB(dbSource)
	return db, err
}
