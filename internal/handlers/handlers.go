package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"yaprfinal/internal/database"
	"yaprfinal/internal/models"
	"yaprfinal/internal/repeat"
)

type Env struct {
	DB *sql.DB
}

// sendError принимает код ответа
func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.Response{Error: msg})
}

func validateTask(t *models.Task) error {
	if t.Title == "" {
		return fmt.Errorf("task header not stated")
	}
	now := time.Now().Truncate(24 * time.Hour)
	if t.Date == "" {
		t.Date = now.Format(repeat.DateLayout)
	}
	dt, err := time.Parse(repeat.DateLayout, t.Date)
	if err != nil {
		return fmt.Errorf("incorrect date format")
	}

	if dt.Before(now) {
		if t.Repeat == "" {
			t.Date = now.Format(repeat.DateLayout)
		} else {
			next, err := repeat.NextDate(now, t.Date, t.Repeat)
			if err != nil {
				return err
			}
			t.Date = next
		}
	} else if t.Repeat != "" {
		if _, err := repeat.NextDate(now, t.Date, t.Repeat); err != nil {
			return err
		}
	}
	return nil
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Добавление задачи (POST)
func (e *Env) addTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		sendError(w, "JSON deserialisation error", http.StatusBadRequest)
		return
	}
	if err := validateTask(&t); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := database.AddTask(e.DB, t)
	if err != nil {
		sendError(w, "write into DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{ID: fmt.Sprint(id)})
}

// Список задач и поиск (GET)
func (e *Env) TasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	if date, err := time.Parse("02.01.2006", search); err == nil {
		search = date.Format(repeat.DateLayout)
	}

	tasks, err := database.GetTasks(e.DB, search)
	if err != nil {
		sendError(w, "DB error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.TasksResponse{Tasks: tasks})
}

// Получение одной задачи (GET)
func (e *Env) getTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		sendError(w, "ID is not stated", http.StatusBadRequest)
		return
	}
	t, err := database.GetTask(e.DB, id)
	if err != nil {
		sendError(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// Обновление задачи (PUT)
func (e *Env) updateTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	json.NewDecoder(r.Body).Decode(&t)
	if t.ID == "" {
		sendError(w, "ID is not stated", http.StatusBadRequest)
		return
	}
	if err := validateTask(&t); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := database.UpdateTask(e.DB, t)
	if err != nil {
		sendError(w, "DB error", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		sendError(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

// Выполнение задачи (POST)
func (e *Env) TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	t, err := database.GetTask(e.DB, id)
	if err != nil {
		sendError(w, "task not found", http.StatusNotFound)
		return
	}
	if t.Repeat == "" {
		database.DeleteTask(e.DB, id)
	} else {
		next, err := repeat.NextDate(time.Now(), t.Date, t.Repeat)
		if err != nil {
			sendError(w, err.Error(), http.StatusBadRequest)
			return
		}
		database.UpdateTaskDate(e.DB, id, next)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

// Удаление задачи (DELETE)
func (e *Env) deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	rows, err := database.DeleteTask(e.DB, id)
	if err != nil {
		sendError(w, "DB error", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		sendError(w, "task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func (e *Env) NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")
	now, err := time.Parse(repeat.DateLayout, nowStr)
	if err != nil {
		sendError(w, "bad now", http.StatusBadRequest)
		return
	}
	next, err := repeat.NextDate(now, dateStr, repeatStr)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(next))
}
