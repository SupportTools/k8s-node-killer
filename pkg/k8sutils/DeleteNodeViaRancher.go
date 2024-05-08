package k8sutils

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/supporttools/k8s-node-killer/pkg/config"
	"k8s.io/client-go/kubernetes"
)

// DeleteNodeViaRancher deletes a node from the Rancher managed cluster based on the node name.
func DeleteNodeViaRancher(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) bool {
	logger.Printf("Starting process to delete node %s via Rancher API...", nodeName)
	logger.Printf("Connecting to Rancher API at: %s", config.CFG.RancherAPI)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	listURL := fmt.Sprintf("%s/v1/cluster.x-k8s.io.machines/fleet-default", config.CFG.RancherAPI)
	logger.Debugf("Generated list URL for machines: %s", listURL)

	req, err := http.NewRequestWithContext(ctx, "GET", listURL, nil)
	if err != nil {
		logger.Errorf("Failed to create GET request for listing machines: %v", err)
		return false
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(config.CFG.RancherKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+authHeader)

	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("Failed to send GET request to list machines: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Failed to read response body from listing machines: %v", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("Failed to list machines, Rancher API responded with status code %d: %s", resp.StatusCode, string(body))
		return false
	}

	var machines MachineList
	if err := json.Unmarshal(body, &machines); err != nil {
		logger.Errorf("Failed to unmarshal machine list response: %v", err)
		return false
	}

	var machineName string
	for _, machine := range machines.Data {
		if machine.Spec.InfrastructureRef.Name == nodeName {
			machineName = machine.Metadata.Name
			break
		}
	}

	if machineName == "" {
		logger.Errorf("No machine found for node name %s", nodeName)
		return false
	}

	deleteURL := fmt.Sprintf("%s/v1/cluster.x-k8s.io.machines/fleet-default/%s", config.CFG.RancherAPI, machineName)
	logger.Debugf("Generated DELETE URL for machine: %s", deleteURL)

	delReq, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		logger.Errorf("Failed to create DELETE request for Rancher API: %v", err)
		return false
	}

	delReq.Header.Set("Content-Type", "application/json")
	delReq.Header.Set("Authorization", "Basic "+authHeader)

	delResp, err := client.Do(delReq)
	if err != nil {
		logger.Errorf("Failed to send DELETE request: %v", err)
		return false
	}
	defer delResp.Body.Close()

	delBody, err := io.ReadAll(delResp.Body)
	if err != nil {
		logger.Errorf("Failed to read delete response body: %v", err)
		return false
	}

	if delResp.StatusCode != http.StatusOK {
		logger.Errorf("Failed to delete machine, Rancher API responded with status code %d: %s", delResp.StatusCode, string(delBody))
		return false
	}

	logger.Infof("Successfully deleted machine for node %s via Rancher API.", nodeName)
	return true
}
