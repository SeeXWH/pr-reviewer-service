package team

import (
	"context"
	"net/http"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/pkg/req"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	teamService *Service
}

func NewHandler(router *http.ServeMux, teamService *Service) {
	handler := &Handler{
		teamService: teamService,
	}
	router.HandleFunc("POST /team/add", handler.Create())
}

func (h *Handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parentCtx := r.Context()
		ctx, cancel := context.WithTimeout(parentCtx, 300*time.Second)
		defer cancel()
		reqBody, err := req.HandleBody[TeamCreateRequestDTO](r)
		if err != nil {
			res.Error(w, 400, "BAD_REQUEST", "bad request")
			return
		}
		teamToCreate := ToDomain(*reqBody)
		createdTeam, err := h.teamService.Create(ctx, &teamToCreate)
		if err != nil {
			switch err {
			case ErrTeamExists:
				res.Error(w, 400, "BAD_REQUEST", err.Error())
			default:
				res.Error(w, 500, "UNKNOWN_ERR", "unknown error")

			}
			return
		}
		response := ToResponse(createdTeam)
		res.JSON(w, 201, response)
	}
}
