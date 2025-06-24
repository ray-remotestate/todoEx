package handlers

import (
	_ "database/sql"
	"encoding/json"
	_ "fmt"
	_ "log"
	"net/http"
	_ "strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/ray-remotestate/todoEx/database"
	"github.com/ray-remotestate/todoEx/middlewares"
	"github.com/ray-remotestate/todoEx/models"
)

/*
func Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
/*
	var task models.Todo
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task.ID = uuid.New()
	task.UserID = user.ID
	task.CreatedAt = time.Now()
	task.Status = "pending"
	task.ArchivedAt = nil

	_, err = database.TodoEx.Exec(`
		INSERT INTO todo (id, user_id, title, description, status, due_date, created_at, archived_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		task.ID, task.UserID, task.Title, task.Description, task.Status, task.DueDate, task.CreatedAt, task.ArchivedAt)

	if err != nil {
		http.Error(w, "failed to insert task into todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}
*/

func Create(w http.ResponseWriter, r *http.Request) {
	user := middlewares.UserContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var task models.Todo
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	task.ID = uuid.New()
	task.UserID = user.ID
	task.CreatedAt = time.Now()
	task.Status = "pending"
	task.ArchivedAt = nil

	_, err = database.TodoEx.Exec(`
		INSERT INTO todo (id, user_id, title, description, status, due_date, created_at, archived_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		task.ID, task.UserID, task.Title, task.Description, task.Status, task.DueDate, task.CreatedAt, task.ArchivedAt)
	if err != nil {
		http.Error(w, "failed to insert task into todo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func Fetch(w http.ResponseWriter, r *http.Request) {
	// authenticate first
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	var userID string
	var expiresAt time.Time

	err := database.TodoEx.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions
		WHERE session_token = $1`, token).Scan(&userID, &expiresAt)

	if err != nil || time.Now().After(expiresAt) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := database.TodoEx.Query(`
		SELECT id, user_id, title, description, status, due_date, created_at, archived_at
		FROM todo
		WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC`, userID)

	if err != nil {
		http.Error(w, "failed to retrieve tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Todo

	for rows.Next() {
		var task models.Todo
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.DueDate, &task.CreatedAt, &task.ArchivedAt)
		if err != nil {
			http.Error(w, "failed to scan task from the database", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	json.NewEncoder(w).Encode(tasks)
}

func Update(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	var userID string
	var expiresAt time.Time

	err := database.TodoEx.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions
		WHERE session_token = $1`, token).Scan(&userID, &expiresAt)

	if err != nil || time.Now().After(expiresAt) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID := vars["id"]
	if taskID == "" {
		http.Error(w, "missing task ID", http.StatusBadRequest)
		return
	}

	var updateTask models.Todo
	err = json.NewDecoder(r.Body).Decode(&updateTask)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, err = database.TodoEx.Exec(`
		UPDATE todo
		SET title = $1, description = $2, status = $3, due_date = $4
		WHERE id = $5 AND user_id = $6 AND archived_at IS NULL`,
		updateTask.Title, updateTask.Description, updateTask.Status, updateTask.DueDate, taskID, userID)

	if err != nil {
		http.Error(w, "failed to update task in the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updateTask)
}

func Archive(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	var userID string
	var expiresAt time.Time

	err := database.TodoEx.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions
		WHERE session_token = $1`, token).Scan(&userID, &expiresAt)

	if err != nil || time.Now().After(expiresAt) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	taskID := vars["id"]
	if taskID == "" {
		http.Error(w, "missing task ID", http.StatusBadRequest)
		return
	}

	_, err = database.TodoEx.Exec(`
		UPDATE todo
		SET archived_at = $1
		WHERE id = $2 AND user_id = $3 AND archived_at IS NULL`,
		time.Now(), taskID, userID)
	if err != nil {
		http.Error(w, "failed to archive task in the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
