package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	_ "github.com/sirupsen/logrus"
	"github.com/ray-remotestate/todoEx/database"
	"github.com/ray-remotestate/todoEx/database/dbHelper"
	"github.com/ray-remotestate/todoEx/middlewares"
	"github.com/ray-remotestate/todoEx/utils"
	"github.com/google/uuid"
)

func Register_Session(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(body.Password) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	exists, err := dbHelper.IsUserExists(body.Email)
	if err != nil {
		http.Error(w, "failed to check user existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(body.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	sessionToken := utils.HashString(body.Email + time.Now().String())
	txErr := database.Tx(func(tx *sql.Tx) error { // using *sql.Tx instead *sql.DB as we want both the operation to either commit together or fail together.
		userID, saveErr := dbHelper.CreateUser(tx, body.Name, body.Email, hashedPassword)
		if saveErr != nil {
			return saveErr
		}
		sessionErr := dbHelper.CreateUserSession(tx, userID, sessionToken)
		if sessionErr != nil {
			return sessionErr
		}
		return nil
	})
	if txErr != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token": sessionToken,
	})
}

func Register_JWT(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(body.Password) < 6 {
		http.Error(w, "password length must be at least 6 characters", http.StatusBadRequest)
		return
	}

	exists, err := dbHelper.IsUserExists(body.Email)
	if err != nil {
		http.Error(w, "failed to check user exixtence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "user already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(body.Password)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	var userID uuid.UUID
	txErr := database.Tx(func(tx *sql.Tx) error {
		var saveErr error
		userID, saveErr = dbHelper.CreateUser(tx, body.Name, body.Email, hashedPassword)
		return saveErr
	})
	if txErr != nil {
		http.Error(w, "failed to register the user", http.StatusInternalServerError)
		return
	}

	JWTToken, err := utils.CreateJWTToken(userID)
	if err != nil {
		http.Error(w, "failed to create jwt token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
		"token": JWTToken,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}

	userID, err := dbHelper.GetUserIDByPassword(body.Email, body.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	sessionToken := utils.HashString(body.Email + time.Now().String())
	err = dbHelper.CreateUserSession(database.TodoEx, userID, sessionToken)
	if err != nil {
		http.Error(w, "failed to create user session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"token": sessionToken,
	})
}

// login using JWT Token
func Login_JWT(w http.ResponseWriter, r *http.Request) {
	body := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid login request", http.StatusBadRequest)
		return
	}

	userID, err := dbHelper.GetUserIDByPassword(body.Email, body.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	JWTToken, err := utils.CreateJWTToken(userID)
	if err != nil {
		http.Error(w, "failed to create JWT token", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Logged in successfully"}`))
	json.NewEncoder(w).Encode(map[string]string{
		"JWTtoken": JWTToken,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	user := middlewares.UserContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		http.Error(w, "missing session token", http.StatusBadRequest)
		return
	}

	dbUserID, err := dbHelper.GetUserIDBySession(token)
	if err != nil || dbUserID != user.ID {
		http.Error(w, "invalid session token", http.StatusUnauthorized)
		return
	}

	_, err = database.TodoEx.Exec(`
		DELETE FROM user_sessions
		WHERE session_token = $1`, token)
	if err != nil {
		http.Error(w, "failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}
