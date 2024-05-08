package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// AppConfig structure for environment-based configurations.
type AppConfig struct {
	Debug                   bool          `json:"debug"`
	MetricsPort             int           `json:"metricsPort"`
	InsecureSkipVerify      bool          `json:"insecureSkipVerify"`
	HarvesterAPI            string        `json:"harvesterAPI"`
	HarvesterKey            string        `json:"harvesterKey"`
	HarvesterNamespace      string        `json:"harvesterNamespace"`
	RancherAPI              string        `json:"rancherAPI"`
	RancherKey              string        `json:"rancherKey"`
	RancherCluster          string        `json:"rancherCluster"`
	RecoveryWaitTimeMinutes int           `json:"recoveryWaitTimeMinutes"`
	DrainTimeoutMinutes     int           `json:"drainTimeoutMinutes"`
	RecoveryDelayMinutes    int           `json:"recoveryDelayMinutes"`
	NewNodeThreshold        time.Duration `json:"newNodeThreshold"`
	RescanInterval          time.Duration `json:"rescanInterval"`
}

var CFG AppConfig

// LoadConfiguration loads configuration from environment variables.
func LoadConfiguration() {
	CFG.Debug = parseEnvBool("DEBUG", false)            // Assuming false as the default value
	CFG.MetricsPort = parseEnvInt("METRICS_PORT", 9090) // Assuming 9090 as the default port
	CFG.InsecureSkipVerify = parseEnvBool("INSECURE_SKIP_VERIFY", false)
	CFG.HarvesterAPI = getEnvOrDefault("HARVESTER_API", "https://harvester.example.com")
	CFG.HarvesterKey = getEnvOrDefault("HARVESTER_KEY", "")
	CFG.HarvesterNamespace = getEnvOrDefault("HARVESTER_NAMESPACE", "default")
	CFG.RancherAPI = getEnvOrDefault("RANCHER_API", "https://rancher.example.com")
	CFG.RancherKey = getEnvOrDefault("RANCHER_KEY", "")
	CFG.RancherCluster = getEnvOrDefault("RANCHER_CLUSTER", "local")
	CFG.RecoveryWaitTimeMinutes = parseEnvInt("RECOVERY_WAIT_TIME_MINUTES", 5)
	CFG.DrainTimeoutMinutes = parseEnvInt("DRAIN_TIMEOUT_MINUTES", 60)
	CFG.RecoveryDelayMinutes = parseEnvInt("RECOVERY_DELAY_MINUTES", 10)
	CFG.NewNodeThreshold = time.Duration(parseEnvInt("NEW_NODE_THRESHOLD", 60)) * time.Minute
	CFG.RescanInterval = time.Duration(parseEnvInt("RESCAN_INTERVAL", 5)) * time.Minute
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Error parsing %s as int: %v. Using default value: %d", key, err, defaultValue)
		return defaultValue
	}
	return intValue
}

func parseEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Error parsing %s as bool: %v. Using default value: %t", key, err, defaultValue)
		return defaultValue
	}
	return boolValue
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port number %d; must be between 1 and 65535", port)
	}
	return nil
}

func validateNonEmpty(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	return nil
}

func ValidateConfiguration(cfg *AppConfig) error {
	if err := validatePort(cfg.MetricsPort); err != nil {
		return err
	}
	if err := validateNonEmpty("harvesterAPI", cfg.HarvesterAPI); err != nil {
		return err
	}
	if err := validateNonEmpty("harvesterKey", cfg.HarvesterKey); err != nil {
		return err
	}
	if err := validateNonEmpty("harvesterNamespace", cfg.HarvesterNamespace); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherAPI", cfg.RancherAPI); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherKey", cfg.RancherKey); err != nil {
		return err
	}
	if err := validateNonEmpty("rancherCluster", cfg.RancherCluster); err != nil {
		return err
	}
	return nil
}
