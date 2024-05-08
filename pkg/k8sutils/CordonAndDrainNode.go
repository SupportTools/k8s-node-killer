package k8sutils

import (
	"context"
	"os"
	"time"

	"github.com/supporttools/k8s-node-killer/pkg/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

func CordonAndDrainNode(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node) bool {
	logger.Printf("Cordoning and draining node %s with a timeout of %d minutes...", node.Name, config.CFG.DrainTimeoutMinutes)
	drainer := &drain.Helper{
		Client:              clientset,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		GracePeriodSeconds:  -1,
		Timeout:             time.Duration(config.CFG.DrainTimeoutMinutes) * time.Minute,
		Out:                 os.Stdout,
		ErrOut:              os.Stderr,
		Ctx:                 ctx,
	}

	if err := drain.RunCordonOrUncordon(drainer, node, true); err != nil {
		logger.Printf("Failed to cordon node %s: %v", node.Name, err)
		return false
	}

	if err := drain.RunNodeDrain(drainer, node.Name); err != nil {
		logger.Printf("Failed to drain node %s: %v", node.Name, err)
		return false
	}

	logger.Printf("Successfully cordoned and drained node %s.", node.Name)
	return true
}
