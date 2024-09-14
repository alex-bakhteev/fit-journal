package main

import (
	"context"
	"fit-journal/internal/config"
	"fit-journal/internal/user"
	"fit-journal/internal/user/db"
	"fit-journal/pkg/client/mongodb"
	"fit-journal/pkg/logging"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("Create router")

	cfg := config.GetConfig()

	mongoDBClient, err := mongodb.NewClient(context.Background(),
		cfg.MongoDB.Host, cfg.MongoDB.Port, cfg.MongoDB.Username, cfg.MongoDB.Password, cfg.MongoDB.Database, cfg.MongoDB.AuthDB)
	if err != nil {
		panic(err)
	}
	//filledUser := user.User{
	//	ID:           "",
	//	Email:        "someone@example.com",
	//	Username:     "Name",
	//	PasswordHash: "12345",
	//}
	storage := db.NewStorage(mongoDBClient, cfg.MongoDB.Collection, logger)
	users, err := storage.FindAll(context.Background())
	fmt.Println(users)

	router := httprouter.New()
	handler := user.NewHandler(logger)
	handler.Register(router)

	start(router, cfg)
}

func start(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()

	var listener net.Listener
	var listenerErr error

	if cfg.Listen.Type == "sock" {
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.Fatal(err)
		}
		socketPath := path.Join(appDir, "app.sock")
		listener, listenerErr = net.Listen("unix", socketPath)
		logger.Infof("Listening on socket: %s", socketPath)
	} else {
		listener, listenerErr = net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Listen.BindIP, cfg.Listen.Port))
		logger.Infof("Start server %s:%s", cfg.Listen.BindIP, cfg.Listen.Port)
	}

	if listenerErr != nil {
		logger.Fatal(listenerErr)
	}

	server := http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second}

	logger.Fatal(server.Serve(listener))
}
