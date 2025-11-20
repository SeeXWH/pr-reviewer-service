package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	userService *Service
}

func NewHandler(router *http.ServeMux, userService *Service) {
	handler := &Handler{
		userService: userService,
	}
	router.HandleFunc("POST /users/setIsActive", handler.UpdateStatus())
}

func (h *Handler) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
		defer cancel()
		var reqBody SetActiveRequestDTO
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			return
		}

		if reqBody.UserID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "user_id is required")
			return
		}

		updatedUser, err := h.userService.SetIsActive(ctx, reqBody.UserID, reqBody.IsActive)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				msg := "User " + reqBody.UserID + " not found"
				res.Error(w, http.StatusNotFound, "NOT_FOUND", msg)
				return
			}
			res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}

		resp := ToResponse(updatedUser)
		res.JSON(w, http.StatusOK, resp)
	}
}
