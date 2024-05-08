package k8sutils

import (
	"context"
	"time"

	"github.com/supporttools/k8s-node-killer/pkg/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForNodeRecovery waits for a node to recover within the specified duration.
func WaitForNodeRecovery(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node) bool {
	totalWaitTime := config.CFG.RecoveryWaitTimeMinutes
	logger.Printf("Starting recovery wait for node %s. Total wait time: %d minutes.", node.Name, totalWaitTime)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for secondsRemaining := totalWaitTime * 60; secondsRemaining > 0; secondsRemaining -= 5 {
		select {
		case <-ctx.Done():
			logger.Printf("Context canceled while waiting for node %s to recover.", node.Name)
			return false
		case <-ticker.C:
			minutes := secondsRemaining / 60
			seconds := secondsRemaining % 60
			ready, err := IsNodeReady(ctx, clientset, node.Name)
			if err != nil {
				logger.Printf("Error checking node readiness: %v", err)
				return false
			}
			if ready {
				logger.Printf("Node %s has recovered after %d minutes and %d seconds.", node.Name, totalWaitTime-minutes, 60-seconds)
				return true
			}
			logger.Printf("Waiting for node %s recovery. %d minutes and %d seconds remaining.", node.Name, minutes, seconds)
		}
	}

	logger.Printf("Node %s did not recover within the allotted %d minutes.", node.Name, totalWaitTime)
	return false
}
