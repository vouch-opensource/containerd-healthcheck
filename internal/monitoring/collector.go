package monitoring

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector represents prom time series exposed to metrics endpoint
type Collector struct {
	healthCheckRequests *prometheus.CounterVec
	healthCheckStatus   *prometheus.GaugeVec
	restartsTotal       prometheus.Counter
}

// NewCollector returns a new prometheus collector
func NewCollector() Collector {
	healthCheckRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "containerd_healthcheck",
			Subsystem: "check",
			Name:      "requests_total",
			Help:      "Current number of health checks performed by container task",
		},
		[]string{"apikey", "statusCode"},
	)

	healthCheckStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "containerd_healthcheck",
			Subsystem: "check",
			Name:      "status",
			Help:      "Current health status of container task",
		},
		[]string{"apikey", "status"},
	)

	restartsTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "containerd_healthcheck",
			Subsystem: "check",
			Name:      "restarts_total",
			Help:      "Total number of restarts performed by containerd-healthcheck",
		},
	)

	return Collector{
		healthCheckRequests: healthCheckRequests,
		healthCheckStatus:   healthCheckStatus,
		restartsTotal:       restartsTotal,
	}

}

// Registry registers the default app collectors
func (c *Collector) Registry() {
	prometheus.MustRegister(c.healthCheckRequests)
	prometheus.MustRegister(c.healthCheckStatus)
	prometheus.MustRegister(c.restartsTotal)
}

// HealthCheckRequests is a prometheus increment function
func (c *Collector) HealthCheckRequests(containerTask string, statusCode int) {
	c.healthCheckRequests.With(prometheus.Labels{"container_task": containerTask, "statusCode": strconv.Itoa(statusCode)}).Inc()
}

// HealthCheckStatus is a prometheus gauge function
func (c *Collector) HealthCheckStatus(containerTask string, status string) {
	c.healthCheckStatus.With(prometheus.Labels{"container_task": containerTask, "status": status})
}

// RestartsTotal is a prometheus increment function
func (c *Collector) RestartsTotal(containerTask string) {
	c.healthCheckStatus.With(prometheus.Labels{"container_task": containerTask})
}
