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

func DrainNode(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node) error {
	logger.Infof("Starting to drain node %s...", node.Name)

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

	err := drain.RunNodeDrain(drainer, node.Name)
	if err != nil {
		logger.Errorf("Failed to drain node %s: %v", node.Name, err)
		return err
	}

	logger.Infof("Successfully drained node %s.", node.Name)
	return nil
}
