package k8sutils

import (
	"bytes"
	"context"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SshAndRebootNode reboots a node by SSHing into it and running the reboot command.
func SshAndRebootNode(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) bool {
	node, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Printf("Failed to retrieve node %s: %v", nodeName, err)
		return false
	}

	if len(node.Status.Addresses) < 1 {
		logger.Printf("No IP address found for node %s, cannot proceed with SSH.", nodeName)
		return false
	}
	nodeIP := node.Status.Addresses[0].Address // Assuming the first address is always the correct one.
	cmd := exec.CommandContext(ctx, "ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "root@"+nodeIP, "uptime; sleep 1; reboot")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Printf("Attempting to reboot node %s via SSH at IP %s...", nodeName, nodeIP)
	if err := cmd.Run(); err != nil {
		logger.Printf("Failed to SSH and reboot node %s: %v", nodeName, err)
		logger.Printf("SSH command output: %s", stdout.String())
		logger.Printf("SSH command error output: %s", stderr.String())
	}
	logger.Printf("Node %s rebooted successfully. SSH output: %s", nodeName, stdout.String())
	return true
}
