package recovery

import (
	"context"
	"time"

	"github.com/supporttools/k8s-node-killer/pkg/health"
	"github.com/supporttools/k8s-node-killer/pkg/k8sutils"
	"github.com/supporttools/k8s-node-killer/pkg/logging"
	"github.com/supporttools/k8s-node-killer/pkg/metrics"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

var logger = logging.SetupLogging()

// AttemptRecovery checks node readiness and performs recovery if necessary.
func AttemptRecovery(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node) {
	overallStartTime := time.Now() // Start timing for overall recovery process

	health.RegisterNodeState(node.Name, "initial_check", "started", "")
	ready, err := k8sutils.IsNodeReady(ctx, clientset, node.Name)
	if err != nil {
		logger.Printf("Error checking node readiness: %v", err)
		return
	}
	if ready {
		health.RegisterNodeState(node.Name, "initial_check", "node_ready", "")
		logger.Printf("Node %s is ready, skipping recovery process.", node.Name)
		return
	}

	health.RegisterNodeState(node.Name, "initial_check", "node_not_ready", "")
	if k8sutils.IsNewNode(node) {
		health.RegisterNodeState(node.Name, "check_new_node", "ignored", "")
		logger.Printf("Node %s is less than an hour old and will be ignored.", node.Name)
		return
	}

	// Define and execute recovery steps
	recoverySteps := map[string]func(context.Context, *kubernetes.Clientset, string) bool{
		"ssh_and_reboot":     k8sutils.SshAndRebootNode,
		"hard_reboot":        k8sutils.HardRebootViaHarvester,
		"delete_via_rancher": k8sutils.DeleteNodeViaRancher,
	}

	allSuccessful := true
	for stepName, stepFunc := range recoverySteps {
		logger.Printf("Starting recovery step '%s' for node %s...", stepName, node.Name)
		stepStartTime := time.Now()
		metrics.RecoveryAttempts.WithLabelValues(node.Name, stepName).Inc()
		health.RegisterNodeState(node.Name, stepName, "in_progress", "")

		stepFunc(ctx, clientset, node.Name) // Execute the recovery step
		if k8sutils.WaitForNodeRecovery(ctx, clientset, node) {
			logger.Printf("Error waiting for node recovery: %v", err)
			allSuccessful = false
			break
		} else if ready {
			metrics.RecoverySuccesses.WithLabelValues(node.Name, stepName).Inc()
			metrics.RecoveryLatencies.WithLabelValues(node.Name, stepName).Observe(time.Since(stepStartTime).Seconds())
			logger.Printf("Recovery step '%s' successful, node %s has recovered.", stepName, node.Name)
			health.RegisterNodeState(node.Name, stepName, "success", "")
			break // Exit the recovery loop if the node has recovered
		} else {
			metrics.RecoveryFailures.WithLabelValues(node.Name, stepName).Inc()
			metrics.RecoveryLatencies.WithLabelValues(node.Name, stepName).Observe(time.Since(stepStartTime).Seconds())
			health.RegisterNodeState(node.Name, stepName, "failure", "")
			allSuccessful = false // Mark as unsuccessful if any step fails
		}
	}

	overallRecoveryDuration := time.Since(overallStartTime)
	metrics.RecoveryTime.WithLabelValues(node.Name).Observe(overallRecoveryDuration.Seconds())
	if !allSuccessful {
		logger.Printf("Failed to fully recover node %s, manual intervention required.", node.Name)
		metrics.NodeDowntime.WithLabelValues(node.Name).Observe(overallRecoveryDuration.Seconds())
		health.RegisterNodeState(node.Name, "overall_recovery", "manual_intervention_required", "")
	}
}
