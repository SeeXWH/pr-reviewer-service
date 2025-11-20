package pullRequest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/SeeXWH/pr-reviewer-service/pkg/req"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	prService PRProvider
}

func NewHandler(router *http.ServeMux, prService PRProvider) {
	handler := Handler{
		prService: prService,
	}

	router.HandleFunc("POST /pullRequest/create", handler.Create())
	router.HandleFunc("POST /pullRequest/merge", handler.Merge())
	router.HandleFunc("POST /pullRequest/reassign", handler.Reassign())
}

func (h *Handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
		defer cancel()

		reqBody, err := req.HandleBody[CreatePRRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			fmt.Println(err.Error())
			return
		}

		if reqBody.PRID == "" || reqBody.AuthorID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "fields required")
			return
		}

		prModel := ToDomain(*reqBody)

		createdPR, err := h.prService.Create(ctx, prModel)
		if err != nil {
			switch {
			case errors.Is(err, ErrAuthorNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
				return
			case errors.Is(err, ErrPRExists):
				res.Error(w, http.StatusConflict, "PR_EXISTS", err.Error())
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}
		resp := ToResponse(createdPR)
		res.JSON(w, http.StatusCreated, resp)
	}
}

func (h *Handler) Merge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
		defer cancel()
		reqBody, err := req.HandleBody[MergePRRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			return
		}

		if reqBody.PRID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "pull_request_id is required")
			return
		}

		mergedPR, err := h.prService.Merge(ctx, reqBody.PRID)
		if err != nil {
			switch {
			case errors.Is(err, ErrPRNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}
		resp := ToResponse(mergedPR)
		res.JSON(w, http.StatusOK, resp)
	}
}

func (h *Handler) Reassign() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 300*time.Millisecond)
		defer cancel()

		reqBody, err := req.HandleBody[ReassignPRRequestDTO](r)
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json")
			return
		}

		if reqBody.PRID == "" || reqBody.OldUserID == "" {
			res.Error(w, http.StatusBadRequest, "BAD_REQUEST", "fields required")
			return
		}

		updatedPR, newReviewer, err := h.prService.ReassignReviewer(ctx, reqBody.PRID, reqBody.OldUserID)
		if err != nil {
			switch {
			case errors.Is(err, ErrPRNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
				return
			case errors.Is(err, ErrPRMerged):
				res.Error(w, http.StatusConflict, "PR_MERGED", err.Error())
				return
			case errors.Is(err, ErrNotAssigned):
				res.Error(w, http.StatusConflict, "NOT_ASSIGNED", err.Error())
				return
			case errors.Is(err, ErrNoCandidate):
				res.Error(w, http.StatusConflict, "NO_CANDIDATE", err.Error())
				return
			default:
				res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
				return
			}
		}

		resp := ToReassignResponse(updatedPR, newReviewer.ID)
		res.JSON(w, http.StatusOK, resp)
	}
}
