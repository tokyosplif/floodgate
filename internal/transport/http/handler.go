package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tokyosplif/floodgate/internal/app/tracking"
	"github.com/tokyosplif/floodgate/internal/domain"
	"github.com/tokyosplif/floodgate/internal/infrastructure/metrics"
)

type ClickService interface {
	HandleClick(ctx context.Context, click domain.Click) error
}

type Handler struct {
	service ClickService
	logger  *slog.Logger
	metrics *metrics.GatewayMetrics
}

func NewHandler(service ClickService, l *slog.Logger, m *metrics.GatewayMetrics) *Handler {
	return &Handler{
		service: service,
		logger:  l,
		metrics: m,
	}
}

func (h *Handler) InitRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Handle("/metrics", promhttp.Handler())
	r.Post("/v1/click", h.ReceiveClick)

	return r
}

func (h *Handler) ReceiveClick(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(h.metrics.RequestDuration.WithLabelValues("ReceiveClick"))
	defer timer.ObserveDuration()

	var c domain.Click
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		h.logger.Error("bad request", slog.Any("err", err))
		h.metrics.ClicksTotal.WithLabelValues("unknown", "error").Inc()
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	c.IP = r.RemoteAddr
	c.UA = r.UserAgent()

	if err := c.Validate(); err != nil {
		h.metrics.ClicksTotal.WithLabelValues(c.CampaignID, "validation_error").Inc()
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.service.HandleClick(r.Context(), c); err != nil {
		if errors.Is(err, tracking.ErrBotDetected) {
			h.metrics.BotsDetected.WithLabelValues("filter_logic").Inc()
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		h.metrics.ClicksTotal.WithLabelValues(c.CampaignID, "internal_error").Inc()
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.metrics.ClicksTotal.WithLabelValues(c.CampaignID, "success").Inc()
	w.WriteHeader(http.StatusAccepted)
}
