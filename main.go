package main

import (
	"log"
	"net/http"
	"os"
	"yaprfinal/internal/database"
	"yaprfinal/internal/handlers"
)

func main() {
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	db, err := database.InitDB(dbFile)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	env := &handlers.Env{DB: db}

	// Обработчики
	http.HandleFunc("/api/task", env.TaskHandler)
	http.HandleFunc("/api/tasks", env.TasksHandler)
	http.HandleFunc("/api/task/done", env.TaskDoneHandler)
	http.HandleFunc("/api/nextdate", env.NextDateHandler)

	// Старт сервера
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Printf("server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
