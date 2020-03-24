package service

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	counterRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests",
		},
		[]string{"handler"},
	)

	counterResponses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "responses",
		},
		[]string{"handler", "code"},
	)

	summaryResponseTime = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "rt",
			Objectives: map[float64]float64{
				0.25: 0.00001,
				0.5:  0.00001,
				0.75: 0.00001,
				0.95: 0.00001,
				0.99: 0.00001,
				1.0:  0.00001,
			},
		},
		[]string{"handler"},
	)
)

func CountRequest(handler string) {
	counterRequests.WithLabelValues(handler).Inc()
}

func CountResponse(handler string, code int) {
	counterResponses.WithLabelValues(handler, strconv.FormatInt(int64(code), 10)).Inc()
}

func CountResponseTime(handler string, duration time.Duration) {
	summaryResponseTime.WithLabelValues(handler).Observe(duration.Seconds())
}
