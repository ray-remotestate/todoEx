package server

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ray-remotestate/todoEx/handlers"
	"github.com/ray-remotestate/todoEx/middlewares"
)

type Server struct {
	Router *mux.Router
	server *http.Server
}

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

func SetupRoutes() *Server {
	router := mux.NewRouter()
	authRoutes := router.PathPrefix("/api").Subrouter()
	authRoutes.Use(middlewares.AuthMiddleware)

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"alive": true}`)
	}).Methods("GET")

	// user
	router.HandleFunc("/register_session", handlers.Register_Session).Methods("POST")
	router.HandleFunc("/register_JWT", handlers.Register_JWT).Methods("POST")
	router.HandleFunc("/login", handlers.Login).Methods("POST")
	authRoutes.HandleFunc("/logout", handlers.Logout).Methods("POST")

	// todo
	authRoutes.HandleFunc("/todos", handlers.Fetch).Methods("GET")
	authRoutes.HandleFunc("/todos", handlers.Create).Methods("POST")
	authRoutes.HandleFunc("/todos/{id}", handlers.Update).Methods("PATCH")
	authRoutes.HandleFunc("/todos/{id}", handlers.Archive).Methods("DELETE")

	return &Server{
		Router: router,
	}
}

func (svr *Server) Run(port string) error {
	svr.server = &http.Server{
		Addr:              port,
		Handler:           svr.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return svr.server.ListenAndServe()
}

func (svr *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return svr.server.Shutdown(ctx)
}
