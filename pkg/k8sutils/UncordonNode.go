package k8sutils

import (
	"context"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
)

// uncordonNode uncordons the given node.
func UncordonNode(ctx context.Context, clientset *kubernetes.Clientset, node *v1.Node) error {
	logger.Printf("Starting to uncordon node %s.", node.Name)

	// Create a drain helper with standard output and error output configurations
	drainer := &drain.Helper{
		Client: clientset,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
		Ctx:    ctx, // Ensure the context is used in the drain operation
	}

	// Attempt to uncordon the node
	logger.Printf("Attempting to uncordon node %s...", node.Name)
	if err := drain.RunCordonOrUncordon(drainer, node, false); err != nil {
		logger.Printf("Failed to uncordon node %s: %v", node.Name, err)
		return err
	}

	// Confirm successful uncordon operation
	logger.Printf("Node %s has been successfully uncordoned.", node.Name)
	return nil
}
