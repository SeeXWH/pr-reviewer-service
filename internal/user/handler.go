package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/pkg/req"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	userService UserProvider
}

func NewHandler(router *http.ServeMux, userService UserProvider) {
	handler := &Handler{
		userService: userService,
	}
	router.HandleFunc("POST /users/setIsActive", handler.UpdateStatus())
	router.HandleFunc("GET /users/getReview", handler.GetReviews())
}

func (h *Handler) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
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
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
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
