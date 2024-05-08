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

func CordonNode(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node, cordon bool) error {
	action := "cordoning"
	if !cordon {
		action = "uncordoning"
	}
	logger.Printf("%s node %s with a timeout of %d minutes...", action, node.Name, config.CFG.DrainTimeoutMinutes)

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

	err := drain.RunCordonOrUncordon(drainer, node, cordon)
	if err != nil {
		logger.Printf("Failed to %s node %s: %v", action, node.Name, err)
		return err
	}

	logger.Printf("Successfully %s node %s.", action, node.Name)
	return nil
}
