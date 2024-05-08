package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/supporttools/k8s-node-killer/pkg/config"
	"github.com/supporttools/k8s-node-killer/pkg/health"
	"github.com/supporttools/k8s-node-killer/pkg/logging"
)

var logger = logging.SetupLogging()

var (
	RecoveryAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_recovery_attempts_total",
		Help: "Total number of recovery attempts by node and step.",
	}, []string{"node", "step"})

	RecoverySuccesses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_recovery_successes_total",
		Help: "Total number of successful recoveries by node and step.",
	}, []string{"node", "step"})

	RecoveryFailures = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_recovery_failures_total",
		Help: "Total number of failed recoveries by node and step.",
	}, []string{"node", "step"})

	RecoveryLatencies = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "k8s_node_killer_recovery_latency_seconds",
		Help:    "Histogram of latencies for recovery actions by node and step.",
		Buckets: prometheus.LinearBuckets(1, 5, 5), // Starting at 1 second, 5 buckets, increment by 5 seconds
	}, []string{"node", "step"})

	RecoveryTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "k8s_node_killer_recovery_time_seconds",
		Help:    "Time taken for the node recovery process, from start to finish.",
		Buckets: prometheus.LinearBuckets(10, 10, 5), // Buckets starting at 10 seconds, incrementing by 10 seconds, with 5 buckets
	}, []string{"node"})

	NodeDowntime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "k8s_node_killer_node_downtime_seconds",
		Help:    "Total downtime for the node during the recovery process.",
		Buckets: prometheus.LinearBuckets(10, 10, 5), // Similar bucket strategy as Recovery Time
	}, []string{"node"})

	InterventionRate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_manual_interventions_total",
		Help: "Total number of times manual intervention was required during recovery.",
	}, []string{"node"})

	RecoveryFailureRate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_recovery_failures_rate",
		Help: "Rate of recovery failures compared to total recovery attempts.",
	}, []string{"node"})

	IncidentFrequency = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_incident_frequency_total",
		Help: "Total number of incidents requiring node recovery.",
	}, []string{"node"})

	ChangeFailureRate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "k8s_node_killer_change_failure_rate",
		Help: "Rate of failures due to changes or updates that required node recovery.",
	}, []string{"node"})
)

func StartMetricsServer() {
	if config.CFG.MetricsPort == 0 {
		logger.Fatalf("Metrics server port not configured")
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/healthz", health.HealthzHandler())
	mux.Handle("/readyz", health.ReadyzHandler())
	mux.Handle("/version", health.VersionHandler())
	mux.HandleFunc("/node-states", health.NodeStatesHandler)

	serverPortStr := strconv.Itoa(config.CFG.MetricsPort)
	logger.Infof("Metrics server starting on port %s", serverPortStr)

	if err := http.ListenAndServe(":"+serverPortStr, mux); err != nil {
		logger.Fatalf("Metrics server failed to start: %v", err)
	}
}
