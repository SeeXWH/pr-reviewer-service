package team

import (
	"context"
	"errors"
	"net/http"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/pkg/req"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	teamService Provider
	conf        *configs.Config
}

func NewHandler(router *http.ServeMux, teamService Provider, conf *configs.Config) {
	handler := &Handler{
		teamService: teamService,
		conf:        conf,
	}
	router.HandleFunc("POST /team/add", handler.Create())
	router.HandleFunc("GET /team/get", handler.Get())
}

func (h *Handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()
		reqBody, err := req.HandleBody[CreateRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "bad request")
			return
		}
		teamToCreate := ToDomain(*reqBody)
		createdTeam, err := h.teamService.Create(ctx, &teamToCreate)
		if err != nil {
			switch {
			case errors.Is(err, ErrTeamExists):
				res.Error(w, http.StatusBadRequest, "TEAM_EXISTS", err.Error())
				return
			default:
				res.Error(w, http.StatusInternalServerError, "UNKNOWN_ERR", "unknown error")
				return
			}
		}
		response := ToResponse(createdTeam)
		res.JSON(w, http.StatusCreated, response)
	}
}

func (h *Handler) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()

		teamName := r.URL.Query().Get("team_name")
		if teamName == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "team_name is required")
			return
		}

		foundTeam, err := h.teamService.GetByName(ctx, teamName)
		if err != nil {
			switch {
			case errors.Is(err, ErrTeamNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}

		resp := ToTeamInfoDTO(foundTeam)
		res.JSON(w, http.StatusOK, resp)
	}
}
