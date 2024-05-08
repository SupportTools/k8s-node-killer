package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/supporttools/k8s-node-killer/pkg/config"
	"github.com/supporttools/k8s-node-killer/pkg/health"
	"github.com/supporttools/k8s-node-killer/pkg/k8sutils"
	"github.com/supporttools/k8s-node-killer/pkg/logging"
	"github.com/supporttools/k8s-node-killer/pkg/metrics"
	"github.com/supporttools/k8s-node-killer/pkg/recovery"
)

var (
	logger       = logging.SetupLogging()
	nodeLocks    = make(map[string]*sync.Mutex)
	mutexMapLock = sync.Mutex{} // Protects access to the nodeLocks map
)

func getNodeMutex(nodeName string) *sync.Mutex {
	mutexMapLock.Lock()
	defer mutexMapLock.Unlock()

	if lock, exists := nodeLocks[nodeName]; exists {
		return lock
	} else {
		nodeLocks[nodeName] = &sync.Mutex{}
		return nodeLocks[nodeName]
	}
}

func scanNodes(ctx context.Context, clientset *kubernetes.Clientset) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Errorf("Failed to list nodes: %v", err)
		return
	}
	for _, node := range nodes.Items {
		logger.Infof("Scanning node %s...", node.Name)
		nodeMutex := getNodeMutex(node.Name)
		nodeMutex.Lock()
		recovery.AttemptRecovery(ctx, clientset, &node)
		nodeMutex.Unlock()
	}
}

func setupSignalHandler(cancelFunc context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Block until a signal is received.
	logger.Println("Shutting down gracefully...")
	cancelFunc()
}

func main() {
	logger.Println("Starting k8s-node-killer...")

	config.LoadConfiguration()
	if err := config.ValidateConfiguration(&config.CFG); err != nil {
		logger.Fatalf("Configuration validation error: %v", err)
	}
	if config.CFG.Debug {
		logger.Println("Debug mode enabled")
		logger.Println("Configuration:")
		logger.Printf(" - Metrics Port: %d", config.CFG.MetricsPort)
		logger.Printf(" - Harvester API: %s", config.CFG.HarvesterAPI)
		logger.Printf(" - Rancher API: %s", config.CFG.RancherAPI)
	}

	logger.Printf("Version: %s", health.Version)
	logger.Printf("Git Commit: %s", health.GitCommit)
	logger.Printf("Build Time: %s", health.BuildTime)

	go func() {
		logger.Println("Starting metrics server...")
		metrics.StartMetricsServer()
	}()

	kubeConfig, err := k8sutils.GetConfig(context.Background())
	if err != nil {
		logger.Fatalf("Error getting Kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatalf("Error creating clientset: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up a signal handler for graceful shutdown
	go setupSignalHandler(cancel)

	nodeInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return clientset.CoreV1().Nodes().List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.CoreV1().Nodes().Watch(ctx, options)
			},
		},
		&v1.Node{},
		0, // Resync period
		cache.Indexers{},
	)

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {},
		UpdateFunc: func(oldObj, newObj interface{}) {
			node := newObj.(*v1.Node)
			nodeMutex := getNodeMutex(node.Name)
			nodeMutex.Lock()
			recovery.AttemptRecovery(ctx, clientset, node)
			nodeMutex.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			if node, ok := obj.(*v1.Node); ok {
				nodeMutex := getNodeMutex(node.Name)
				nodeMutex.Lock()
				defer nodeMutex.Unlock()
				delete(nodeLocks, node.Name)
			}
		},
	})

	go nodeInformer.Run(ctx.Done())

	// Initial scan of all nodes
	scanNodes(ctx, clientset)

	// Start periodic checks
	ticker := time.NewTicker(config.CFG.RescanInterval * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				scanNodes(ctx, clientset)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	select {}
}
