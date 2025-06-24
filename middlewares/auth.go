package middlewares

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/ray-remotestate/todoEx/database/dbHelper"
	"github.com/ray-remotestate/todoEx/models"
	"github.com/sirupsen/logrus"
)

type ContextKeys string

const (
	userContext ContextKeys = "__userContext"
)

func UserContext(r *http.Request) *models.User {
	if user, ok := r.Context().Value(userContext).(*models.User); ok && user != nil {
		return user
	}
	return nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Get user_id from session token
		userID, err := dbHelper.GetUserIDBySession(token)
		if err != nil {
			logrus.Printf("%v", err)
			http.Error(w, "invalid or expired session token", http.StatusUnauthorized)
			return
		}

		// Get user details
		user, err := dbHelper.GetUserByUserID(userID)
		if err == sql.ErrNoRows {
			http.Error(w, "user does not exist", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), userContext, &user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
