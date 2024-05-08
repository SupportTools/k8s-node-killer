package logging

import (
	v1 "k8s.io/api/core/v1"
)

func LogNodeDetails(node *v1.Node) {
	logger.Printf("Logging details for node: %s\n", node.Name)
	logger.Printf("Node UID: %s\n", node.UID)
	logger.Printf("Node Status: %s\n", node.Status.Phase)
	logger.Printf("Node Conditions:\n")
	for _, cond := range node.Status.Conditions {
		logger.Printf(" - Type: %s, Status: %s, LastHeartbeatTime: %s, LastTransitionTime: %s, Reason: %s, Message: %s\n",
			cond.Type, cond.Status, cond.LastHeartbeatTime, cond.LastTransitionTime, cond.Reason, cond.Message)
	}
	logger.Printf("Node Addresses:\n")
	for _, address := range node.Status.Addresses {
		logger.Printf(" - Type: %s, Address: %s\n", address.Type, address.Address)
	}
	logger.Printf("Node Labels:\n")
	for key, value := range node.Labels {
		logger.Printf(" - %s: %s\n", key, value)
	}
	logger.Printf("Node Annotations:\n")
	for key, value := range node.Annotations {
		logger.Printf(" - %s: %s\n", key, value)
	}
	logger.Printf("Node Taints:\n")
	for _, taint := range node.Spec.Taints {
		logger.Printf(" - Key: %s, Value: %s, Effect: %s\n", taint.Key, taint.Value, taint.Effect)
	}
	logger.Printf("Node Creation Timestamp: %s\n", node.CreationTimestamp)
}
