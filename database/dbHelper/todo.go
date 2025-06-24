package dbHelper

// "net/http"
// "database/sql"
// "time"

// "github.com/google/uuid"

// "github.com/ray-remotestate/todoEx/database"
// "github.com/ray-remotestate/todoEx/models"

// func CreateTodo(task models.Todo, user *models.User) (uuid.UUID, error){

// 	task.ID = uuid.New()
// 	task.UserID = user.ID
// 	task.CreatedAt = time.Now()
// 	task.Status = "pending"
// 	task.ArchivedAt = nil

// 	_, err = database.TodoEx.Exec(`
// 		INSERT INTO todo (id, user_id, title, description, status, due_date, created_at, archived_at)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
// 		task.ID, task.UserID, task.Title, task.Description, task.Status, task.DueDate, task.CreatedAt, task.ArchivedAt)
// 	if err != nil {
// 		http.Error(w, "failed to insert task into todo", http.StatusInternalServerError)
// 		return
// 	}
// }
