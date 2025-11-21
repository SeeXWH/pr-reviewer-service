package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/pkg/req"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	userService Provider
	conf        *configs.Config
}

func NewHandler(router *http.ServeMux, userService Provider, conf *configs.Config) {
	handler := &Handler{
		userService: userService,
		conf:        conf,
	}
	router.HandleFunc("POST /users/setIsActive", handler.UpdateStatus())
	router.HandleFunc("GET /users/getReview", handler.GetReviews())
	router.HandleFunc("POST /user/massDeactivate", handler.MassDeactivate())
}

func (h *Handler) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()
		reqBody, err := req.HandleBody[SetActiveRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			return
		}

		if reqBody.UserID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}

		updatedUser, err := h.userService.SetIsActive(ctx, reqBody.UserID, reqBody.IsActive)
		if err != nil {
			switch {
			case errors.Is(err, ErrUserNotFound):
				msg := "User " + reqBody.UserID + " not found"
				res.Error(w, http.StatusNotFound, "NOT_FOUND", msg)
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}

		resp := ToResponse(updatedUser)
		res.JSON(w, http.StatusOK, resp)
	}
}

func (h *Handler) GetReviews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()

		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}

		prModels, err := h.userService.GetReviews(ctx, userID)
		if err != nil {
			switch {
			case errors.Is(err, ErrUserNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", "User "+userID+" not found")
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}

		resp := ToReviewsResponse(userID, prModels)
		res.JSON(w, http.StatusOK, resp)
	}
}

func (h *Handler) MassDeactivate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()
		reqBody, err := req.HandleBody[MassDeactivateRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			return
		}
		if reqBody.TeamName == "" || len(reqBody.UserIDs) == 0 {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "team_name and user_ids are required")
			return
		}
		result, err := h.userService.MassDeactivate(ctx, reqBody.TeamName, reqBody.UserIDs)
		if err != nil {
			res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		resp := ToMassDeactivateResponse(result)
		res.JSON(w, http.StatusOK, resp)
	}
}
