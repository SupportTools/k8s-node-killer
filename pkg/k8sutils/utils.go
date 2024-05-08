package k8sutils

import (
	"github.com/supporttools/k8s-node-killer/pkg/logging"
)

var logger = logging.SetupLogging()

// MachineList represents the list of machines fetched from the Rancher API
type MachineList struct {
	Data []Machine `json:"data"`
}

// Machine represents a single machine in the Rancher API response
type Machine struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Spec struct {
		InfrastructureRef struct {
			Name string `json:"name"`
		} `json:"infrastructureRef"`
	} `json:"spec"`
}
