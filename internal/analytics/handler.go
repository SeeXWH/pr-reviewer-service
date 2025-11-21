package analytics

import (
	"context"
	"net/http"

	"github.com/SeeXWH/pr-reviewer-service/configs"
	"github.com/SeeXWH/pr-reviewer-service/pkg/res"
)

type Handler struct {
	analyticService Provider
	conf            *configs.Config
}

func NewHandler(router *http.ServeMux, analyticService Provider, conf *configs.Config) {
	handler := &Handler{
		analyticService: analyticService,
		conf:            conf,
	}

	router.HandleFunc("GET /analytics/pr", handler.GetStats())
}

func (h *Handler) GetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.conf.App.TimeOut)
		defer cancel()
		data, err := h.analyticService.GetStats(ctx)
		if err != nil {
			res.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "unknown error")
			return
		}
		resp := ToDTO(data)
		res.JSON(w, http.StatusOK, resp)
	}
}
