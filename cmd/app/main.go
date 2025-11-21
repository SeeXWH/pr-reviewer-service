package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/internal/pullrequest"
	"github.com/SeeXWH/pr-reviewer-service/internal/team"
	"github.com/SeeXWH/pr-reviewer-service/internal/user"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
	"github.com/SeeXWH/pr-reviewer-service/pkg/logger"
	"github.com/SeeXWH/pr-reviewer-service/pkg/middleware"
)

func main() {
	log := logger.Setup()
	slog.SetDefault(log)

	start := time.Now()

	conf := configs.Load()
	log.Info("config loaded", "db_host", conf.DB.Host, "db_port", conf.DB.Port)

	postgresDb := db.NewPostgresDb(conf)
	mainRouter := http.NewServeMux()

	teamRepository := team.NewRepository(postgresDb)
	userRepository := user.NewRepository(postgresDb)
	prRepository := pullrequest.NewRepository(postgresDb)

	teamService := team.NewService(teamRepository, log)
	userService := user.NewService(userRepository, log)
	prService := pullrequest.NewService(userService, prRepository, log)

	user.NewHandler(mainRouter, userService)
	team.NewHandler(mainRouter, teamService)
	pullrequest.NewHandler(mainRouter, prService)

	server := http.Server{
		Addr:    conf.App.Port,
		Handler: middleware.Logging(mainRouter),
	}
	go func() {
		log.Info("server started",
			"address", conf.App.Port,
			"startup_duration", time.Since(start).String(),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}
	log.Info("Server exited properly")
}
