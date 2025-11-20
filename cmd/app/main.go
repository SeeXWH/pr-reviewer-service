package main

import (
	"context"
	"errors"
	"fmt"
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
	now := time.Now()
	conf := configs.Load()
	fmt.Println(conf.DB.Host)
	db := db.NewPostgresDb(conf)
	mainRouter := http.NewServeMux()

	teamRepository := team.NewRepository(db)
	userRepository := user.NewRepository(db)
	prRepository := pullrequest.NewRepository(db)

	teamService := team.NewService(teamRepository)
	userService := user.NewService(userRepository)
	prService := pullrequest.NewService(userService, prRepository)

	user.NewHandler(mainRouter, userService)
	team.NewHandler(mainRouter, teamService)
	pullrequest.NewHandler(mainRouter, prService)

	server := http.Server{
		Addr:    ":8080",
		Handler: middleware.Logging(mainRouter),
	}
	go func() {
		log.Info("Server listening on :8080 and start in " + time.Since(now).String())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
	}
	log.Info("Server exited properly")
}
