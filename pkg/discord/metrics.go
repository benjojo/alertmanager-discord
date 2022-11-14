package discord

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsToDiscordInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "discord_client_requests_in_flight",
		Help: "The current number of http requests being sent by the discord client.",
	})

	RequestsToDiscordTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "discord_client_requests_total",
		Help: "The total number of http requests sent by the discord client.",
	}, []string{"code", "method"})

	RequestsToDiscordDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "discord_client_request_duration_seconds",
		Help:    "Duration of all http requests sent by the discord client.",
		Buckets: prometheus.DefBuckets,
	}, []string{"code"})
)
