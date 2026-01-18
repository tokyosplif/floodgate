package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type GatewayMetrics struct {
	ClicksTotal     *prometheus.CounterVec
	BotsDetected    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

func NewGatewayMetrics() *GatewayMetrics {
	return &GatewayMetrics{
		ClicksTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "floodgate_clicks_total",
			Help: "Total number of processed clicks",
		}, []string{"campaign_id", "status"}),

		BotsDetected: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "floodgate_bots_total",
			Help: "Total number of detected bots",
		}, []string{"reason"}),

		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "floodgate_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"handler"}),
	}
}
