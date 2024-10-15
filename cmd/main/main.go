package main

import (
	"context"
	"fit-journal/internal/config"
	user "fit-journal/internal/entities/user"
	userDB "fit-journal/internal/entities/user/db"
	workout "fit-journal/internal/entities/workout"
	"fit-journal/internal/entities/workout/db"
	"fit-journal/pkg/client/postgresql"
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
	router := httprouter.New() // Необходимо инициализировать роутер

	// Получение конфигурации
	cfg := config.GetConfig()

	// Инициализация PostgreSQL клиента
	logger.Info("Initialize PostgreSQL client")
	ctx := context.Background()
	pgClient, err := postgresql.NewClient(ctx, 3, cfg.Storage)
	if err != nil {
		logger.Fatalf("Failed to initialize PostgreSQL client: %v", err)
	}
	defer pgClient.Close()

	// Регистрируем репозиторий для пользователя
	logger.Info("Initialize user repository")
	userRepo := userDB.NewRepository(pgClient, logger)

	// Регистрируем хендлеры для пользователя
	logger.Info("Register user handler")
	userHandler := user.NewHandler(logger, userRepo)
	userHandler.Register(router)

	workoutRepo := db.NewRepository(pgClient, logger)
	workoutHandler := workout.NewHandler(logger, workoutRepo, userRepo)
	workoutHandler.Register(router)

	// Запускаем сервер
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
		ReadTimeout:  15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}
