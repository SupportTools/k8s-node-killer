package k8sutils

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/supporttools/k8s-node-killer/pkg/config"
	"k8s.io/client-go/kubernetes"
)

// HardRebootViaHarvester reboots a virtual machine managed by Harvester via an API call.
func HardRebootViaHarvester(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) bool {
	url := fmt.Sprintf("%s/v1/harvester/kubevirt.io.virtualmachines/%s/%s?action=restart", config.CFG.HarvesterAPI, config.CFG.HarvesterNamespace, nodeName)
	logger.Printf("Preparing to send reboot request to URL: %s", url)

	// Configure the client to ignore certificate validation if needed
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.CFG.InsecureSkipVerify}, // Using configuration to manage certificate checks
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logger.Printf("Failed to create request for Harvester API: %v", err)
		return false
	}

	req.Header.Add("Authorization", "Bearer "+config.CFG.HarvesterKey)
	req.Header.Add("Content-Type", "application/json")

	logger.Printf("Sending reboot request for node %s...", nodeName)
	resp, err := client.Do(req)
	if err != nil {
		logger.Printf("Failed to send request to Harvester API: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Failed to read response body from Harvester API: %v", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		logger.Printf("Harvester API responded with status code %d: %s", resp.StatusCode, string(body))
		return false
	}

	logger.Printf("Successfully triggered reboot for VM %s via Harvester API.", nodeName)
	return true
}
