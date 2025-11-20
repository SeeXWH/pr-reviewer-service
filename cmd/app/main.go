package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/internal/team"
	"github.com/SeeXWH/pr-reviewer-service/internal/user"
	"github.com/SeeXWH/pr-reviewer-service/pkg/db"
)

func main() {
	now := time.Now()
	conf := configs.Load()
	fmt.Println(conf.DB.Host)
	db := db.NewPostgresDb(conf)
	mainRouter := http.NewServeMux()

	teamRepository := team.NewRepository(db)
	userRepository := user.NewRepository(db)

	teamService := team.NewService(teamRepository)
	userService := user.NewService(userRepository)

	user.NewHandler(mainRouter, userService)
	team.NewHandler(mainRouter, teamService)

	server := http.Server{
		Addr:    ":8080",
		Handler: mainRouter,
	}
	log.Printf("Server start on %s port. Time: %.3fs\n", server.Addr, time.Since(now).Seconds())
	server.ListenAndServe()
}
