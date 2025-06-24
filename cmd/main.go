package main

import (
	_ "encoding/json"
	"log"
	_ "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ray-remotestate/todoEx/database"
	"github.com/ray-remotestate/todoEx/server"
	"github.com/ray-remotestate/todoEx/config"
)

const shutDownTimeOut = 10 * time.Second

func main() {
	config.Init()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	svr := server.SetupRoutes()

	if err := database.ConnectAndMigrate(); err != nil {
		logrus.Panicf("Failed to initialize the database with error: %v", err)
	}
	logrus.Println("Migration is successful")

	go func() {
		log.Println("Server starting at :8080")
		if err := svr.Run(":8080"); err != nil {
			logrus.Panicf("Server didn't start! %+v", err)
		}
	}()

	<-done

	logrus.Info("Shutting down server...")
	if err := database.ShutdownDatabase(); err != nil {
		logrus.WithError(err).Error("Failed to close database connection")
	}
	if err := svr.Shutdown(shutDownTimeOut); err != nil {
		logrus.WithError(err).Panic("failed to gracefully shutdown server")
	}
	logrus.Info("System is shut. Run again.")
}

/*
	1. Create Server
	2. Routing
	3. Connect to DB
	4. Migration
	5. Run the server
	6. Shut down the server
*/
