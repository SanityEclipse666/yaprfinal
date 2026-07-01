package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"yaprfinal/internal/models"
	"yaprfinal/internal/repeat"
)

type Env struct {
	DB *sql.DB
}

func (e *Env) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		e.getTask(w, r)
	case http.MethodPost:
		e.addTask(w, r)
	case http.MethodPut:
		e.updateTask(w, r)
	case http.MethodDelete:
		e.deleteTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Добавление задачи (POST)
func (e *Env) addTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		sendError(w, "JSON deserialisation error")
		return
	}

	if err := validateTask(&t); err != nil {
		sendError(w, err.Error())
		return
	}

	res, err := e.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		t.Date, t.Title, t.Comment, t.Repeat)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	id, _ := res.LastInsertId()
	json.NewEncoder(w).Encode(models.Response{ID: strconv.FormatInt(id, 10)})
}

// Список задач и поиск (GET)
func (e *Env) TasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	var rows *sql.Rows
	var err error

	if search == "" {
		rows, err = e.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50")
	} else {
		if date, errDate := time.Parse("02.01.2006", search); errDate == nil {
			rows, err = e.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT 50", date.Format(repeat.DateLayout))
		} else {
			query := "%" + search + "%"
			rows, err = e.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT 50", query, query)
		}
	}

	if err != nil {
		sendError(w, err.Error())
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var t models.Task
		var id int64
		rows.Scan(&id, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		t.ID = strconv.FormatInt(id, 10)
		tasks = append(tasks, t)
	}
	json.NewEncoder(w).Encode(models.TasksResponse{Tasks: tasks})
}

// Получение одной задачи (GET)
func (e *Env) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	row := e.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	var t models.Task
	var idInt int64
	if err := row.Scan(&idInt, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
		sendError(w, "task not found")
		return
	}
	t.ID = strconv.FormatInt(idInt, 10)
	json.NewEncoder(w).Encode(t)
}

// Получение следующей даты (GET)
func (e *Env) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	now, err := time.Parse(repeat.DateLayout, nowStr)
	if err != nil {
		http.Error(w, "invalid now", http.StatusBadRequest)
		return
	}

	next, err := repeat.NextDate(now, dateStr, repeatStr)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(next))
}

// Обновление задачи (PUT)
func (e *Env) updateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		sendError(w, "JSON deserialisation error")
		return
	}

	if t.ID == "" {
		sendError(w, "ID not stated")
		return
	}

	if err := validateTask(&t); err != nil {
		sendError(w, err.Error())
		return
	}

	res, err := e.DB.Exec("UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?",
		t.Date, t.Title, t.Comment, t.Repeat, t.ID)

	if err != nil {
		sendError(w, err.Error())
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		sendError(w, "task not found")
		return
	}
	w.Write([]byte("{}"))
}

// Выполнение задачи (POST)
func (e *Env) TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "ID not stated")
		return
	}

	var t models.Task
	err := e.DB.QueryRow("SELECT date, repeat FROM scheduler WHERE id=?", id).Scan(&t.Date, &t.Repeat)
	if err != nil {
		sendError(w, "task not found")
		return
	}

	if t.Repeat == "" {
		// Удаляем одноразовую
		e.DB.Exec("DELETE FROM scheduler WHERE id=?", id)
	} else {

		now := time.Now()
		next, err := repeat.NextDate(now, t.Date, t.Repeat)
		if err != nil {
			sendError(w, "data calculation error: "+err.Error())
			return
		}
		_, err = e.DB.Exec("UPDATE scheduler SET date=? WHERE id=?", next, id)
		if err != nil {
			sendError(w, "DB error")
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

// Удаление задачи (DELETE)
func (e *Env) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "ID not stated")
		return
	}

	res, err := e.DB.Exec("DELETE FROM scheduler WHERE id=?", id)
	if err != nil {
		sendError(w, err.Error())
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		sendError(w, "task not found")
		return
	}
	w.Write([]byte("{}"))
}

// Валидация
func validateTask(t *models.Task) error {
	if t.Title == "" {
		return fmt.Errorf("task header not stated")
	}

	// Получаем текущую дату (полночь локального времени)
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if t.Date == "" {
		t.Date = now.Format(repeat.DateLayout)
	}

	dt, err := time.ParseInLocation(repeat.DateLayout, t.Date, now.Location())
	if err != nil {
		return fmt.Errorf("invalid data format")
	}

	// Если дата в прошлом
	if dt.Before(now) {
		if t.Repeat == "" {
			t.Date = now.Format(repeat.DateLayout)
		} else {
			// Вычисляем следующую дату, так как текущая в прошлом
			next, err := repeat.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return err
			}
			t.Date = next
		}
	} else {

		if t.Repeat != "" {
			_, err := repeat.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Доп функция для обработки ошибок
func sendError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	resp := models.Response{
		Error: msg,
	}

	json.NewEncoder(w).Encode(resp)
}
