package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "circuit"

var (
	once sync.Once

	requestOutcomes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "interceptor_outcomes_total",
			Help:      "gRPC client CB interceptor: per-method outcomes (success, open, throttled, upstream_error).",
		},
		[]string{"name", "outcome"},
	)

	stateTransitions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "state_transitions_total",
			Help:      "gobreaker state transitions; from/to as gobreaker.State.String().",
		},
		[]string{"name", "from", "to"},
	)

	breakerOpen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "open",
			Help:      "1 if circuit is open, 0 if closed or half-open.",
		},
		[]string{"name"},
	)
)

func init() {
	once.Do(func() {
		prometheus.MustRegister(
			requestOutcomes,
			stateTransitions,
			breakerOpen,
		)
	})
}

func IncRequestOutcome(name, outcome string) {
	requestOutcomes.WithLabelValues(name, outcome).Inc()
}

func IncStateTransition(name, from, to string) {
	stateTransitions.WithLabelValues(name, from, to).Inc()
}

func SetBreakerOpen(name string, isOpen bool) {
	v := 0.0
	if isOpen {
		v = 1.0
	}
	breakerOpen.WithLabelValues(name).Set(v)
}
