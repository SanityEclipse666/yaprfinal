package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Код создания таблицы
const (
	tableSchema = `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
		title VARCHAR(255) NOT NULL DEFAULT "",
		comment TEXT NOT NULL DEFAULT "",
		repeat VARCHAR(128) NOT NULL DEFAULT ""
	);`

	indexSchema = `CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);`
)

func InitDB(dbFile string) (*sql.DB, error) {
	// Открытие БД
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Создание таблицы
	if _, err := db.Exec(tableSchema); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Создание индекса
	if _, err := db.Exec(indexSchema); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return db, nil
}
