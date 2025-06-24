package dbHelper

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/ray-remotestate/todoEx/database"
	"github.com/ray-remotestate/todoEx/models"
)

type SQLExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func CreateUser(tx *sql.Tx, name, email, hashedPassword string) (uuid.UUID, error) {
	id := uuid.New()
	createdAt := time.Now().UTC()
	_, err := tx.Exec(`INSERT INTO users (id, name, email, password, created_at) VALUES ($1, $2, $3, $4, $5)`,
		id, name, email, hashedPassword, createdAt)
	return id, err
}

func IsUserExists(email string) (bool, error) {
	var count int
	err := database.TodoEx.QueryRow(`SELECT COUNT(*) FROM users WHERE LOWER(email) = LOWER($1)`, email).Scan(&count)
	return count > 0, err
}

func CreateUserSession(exec SQLExecutor, userID uuid.UUID, token string) error {
	expiresAt := time.Now().Add(120 * time.Hour)
	_, err := exec.Exec(`
		INSERT INTO user_sessions (id, user_id, session_token, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(), userID, token, time.Now(), expiresAt)
	return err
}

func GetUserIDByPassword(email, password string) (uuid.UUID, error) {
	var id uuid.UUID
	var hashedPassword string

	err := database.TodoEx.QueryRow(`
		SELECT id, password FROM users 
		WHERE LOWER(email) = LOWER($1) AND archived_at IS NULL`, email).
		Scan(&id, &hashedPassword)
	if err != nil {
		return uuid.Nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		return uuid.Nil, fmt.Errorf("incorrect password")
	}

	return id, nil
}

func GetUserIDBySession(sessionToken string) (uuid.UUID, error) {
	var userID uuid.UUID

	err := database.TodoEx.QueryRow(`
		SELECT user_id FROM user_sessions
		WHERE session_token = $1 AND expires_at > NOW()`, sessionToken).Scan(&userID)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetUserByUserID(userID uuid.UUID) (models.User, error) {
	var user models.User

	err := database.TodoEx.QueryRow(`
		SELECT id, name, email, password, created_at, archived_at FROM users
		WHERE id = $1 AND archived_at IS NULL`, userID).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.ArchivedAt)
	if err != nil {
		logrus.Printf("%v", err) // remove later (just debugging)
		return models.User{}, err
	}

	return user, nil
}
