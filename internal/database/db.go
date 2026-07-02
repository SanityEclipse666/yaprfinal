package database

import (
	"database/sql"
	"fmt"
	"yaprfinal/internal/models"

	_ "modernc.org/sqlite"
)

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
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(tableSchema); err != nil {
		return nil, err
	}
	if _, err := db.Exec(indexSchema); err != nil {
		return nil, err
	}
	return db, nil
}

// AddTask добавляет задачу и возвращает её ID
func AddTask(db *sql.DB, t models.Task) (int64, error) {
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTasks получает список задач с фильтром по поиску
func GetTasks(db *sql.DB, search string) ([]models.Task, error) {
	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50")
	} else {
		query := "%" + search + "%"
		rows, err = db.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? OR date = ? ORDER BY date LIMIT 50",
			query, query, search)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var id int64
		if err := rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		t.ID = fmt.Sprint(id)
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil

}

// GetTask получает одну задачу по ID
func GetTask(db *sql.DB, id string) (models.Task, error) {
	var t models.Task
	var idInt int64
	err := db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		return t, err
	}
	t.ID = fmt.Sprint(idInt)
	return t, nil
}

// UpdateTask обновляет поля задачи
func UpdateTask(db *sql.DB, t models.Task) (int64, error) {
	res, err := db.Exec("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?",
		t.Date, t.Title, t.Comment, t.Repeat, t.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// DeleteTask удаляет задачу
func DeleteTask(db *sql.DB, id string) (int64, error) {
	res, err := db.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateTaskDate обновляет дату
func UpdateTaskDate(db *sql.DB, id string, nextDate string) error {
	_, err := db.Exec("UPDATE scheduler SET date=? WHERE id=?", nextDate, id)
	return err
}
