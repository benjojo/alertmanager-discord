package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsToAlertForwarderInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "alertmanager_discord_alert_forwarder_requests_in_flight",
		Help: "The current number of events being processed by alert forwarder.",
	})

	RequestsToAlertForwarderTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alertmanager_discord_alert_forwarder_requests_total",
		Help: "The total number of http requests processed by alert forwarder.",
	}, []string{"code"})

	RequestsToAlertForwarderDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "alertmanager_discord_alert_forwarder_request_duration_seconds",
		Help:    "Duration of all http requests processed by alert forwarder.",
		Buckets: prometheus.DefBuckets,
	}, []string{"code"})
)
