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
	"github.com/SeeXWH/pr-reviewer-service/internal/analytics"
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

	postgresDB, err := db.NewPostgresDB(conf)
	if err != nil {
		log.Warn("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	mainRouter := http.NewServeMux()

	teamRepository := team.NewRepository(postgresDB)
	userRepository := user.NewRepository(postgresDB)
	prRepository := pullrequest.NewRepository(postgresDB)
	analyticRepository := analytics.NewRepository(postgresDB)

	teamService := team.NewService(teamRepository, log)
	userService := user.NewService(userRepository, log)
	prService := pullrequest.NewService(userService, prRepository, log)
	analyticsService := analytics.NewService(analyticRepository, log)

	user.NewHandler(mainRouter, userService, conf)
	team.NewHandler(mainRouter, teamService, conf)
	pullrequest.NewHandler(mainRouter, prService, conf)
	analytics.NewHandler(mainRouter, analyticsService, conf)

	server := http.Server{
		Addr:              conf.App.Port,
		Handler:           middleware.Logging(mainRouter),
		ReadHeaderTimeout: conf.App.TimeOut,
	}
	go func() {
		log.Info("server started",
			"address", conf.App.Port,
			"startup_duration", time.Since(start).String(),
		)
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	if err = server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}
	log.Info("Server exited properly")
}
