package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	inClusterModeInformer bool
	kubeconfigInformer    string
	namespaceInformer     string
	resyncPeriodInformer  time.Duration
)

// InformerCmd represents the informer command
var InformerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Watch Kubernetes deployments using informer",
	Long: `Watch Kubernetes deployments using a list/watch informer.
This command connects to a Kubernetes cluster and watches for deployment events.
It supports both in-cluster authentication and kubeconfig-based authentication.`,
	Run: runInformer,
}

func init() {
	// Add flags for authentication and configuration
	InformerCmd.Flags().BoolVar(&inClusterModeInformer, "in-cluster", false, "Use in-cluster authentication (default: false)")
	InformerCmd.Flags().StringVar(&kubeconfigInformer, "kubeconfig", "", "Path to kubeconfig file (default: $HOME/.kube/config)")
	InformerCmd.Flags().StringVar(&namespaceInformer, "namespace", "", "Namespace to watch (default: all namespaces)")
	InformerCmd.Flags().DurationVar(&resyncPeriodInformer, "resync-period", 10*time.Minute, "Resync period for the informer")

	// Add the informer command to root
	rootCmd.AddCommand(InformerCmd)
}

func runInformer(cmd *cobra.Command, args []string) {
	log.Info().Msg("Starting Kubernetes deployment informer")

	// Create Kubernetes client
	clientset, err := createKubernetesClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Create deployment informer
	informer := createDeploymentInformer(clientset)

	// Start the informer
	stopCh := make(chan struct{})
	defer close(stopCh)

	log.Info().Msg("Starting informer...")
	go informer.Run(stopCh)

	// Wait for informer to sync
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		log.Fatal().Msg("Failed to sync informer cache")
	}

	log.Info().Msg("Informer cache synced successfully")

	// Keep the informer running
	select {}
}

func createKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if inClusterModeInformer {
		log.Info().Msg("Using in-cluster authentication")
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	} else {
		log.Info().Msg("Using kubeconfig authentication")
		if kubeconfigInformer == "" {
			home := os.Getenv("HOME")
			if home == "" {
				home = os.Getenv("USERPROFILE") // Windows
			}
			kubeconfigInformer = filepath.Join(home, ".kube", "config")
		}

		log.Info().Str("kubeconfig", kubeconfigInformer).Msg("Loading kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigInformer)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Test the connection
	_, err = clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	log.Info().Msg("Successfully connected to Kubernetes cluster")
	return clientset, nil
}

func createDeploymentInformer(clientset *kubernetes.Clientset) cache.SharedIndexInformer {
	// Create list watcher for deployments
	var listWatcher cache.ListerWatcher
	if namespaceInformer == "" {
		listWatcher = cache.NewListWatchFromClient(
			clientset.AppsV1().RESTClient(),
			"deployments",
			metav1.NamespaceAll,
			fields.Everything(),
		)
		log.Info().Msg("Watching deployments in all namespaces")
	} else {
		listWatcher = cache.NewListWatchFromClient(
			clientset.AppsV1().RESTClient(),
			"deployments",
			namespaceInformer,
			fields.Everything(),
		)
		log.Info().Str("namespace", namespaceInformer).Msg("Watching deployments in namespace")
	}

	// Create informer
	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&appsv1.Deployment{},
		resyncPeriodInformer,
		cache.Indexers{},
	)

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			log.Info().
				Str("event", "ADDED").
				Str("name", deployment.Name).
				Str("namespace", deployment.Namespace).
				Int32("replicas", *deployment.Spec.Replicas).
				Msg("Deployment added")
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldDeployment := oldObj.(*appsv1.Deployment)
			newDeployment := newObj.(*appsv1.Deployment)

			log.Info().
				Str("event", "UPDATED").
				Str("name", newDeployment.Name).
				Str("namespace", newDeployment.Namespace).
				Int32("old-replicas", *oldDeployment.Spec.Replicas).
				Int32("new-replicas", *newDeployment.Spec.Replicas).
				Str("resource-version", newDeployment.ResourceVersion).
				Msg("Deployment updated")
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			log.Info().
				Str("event", "DELETED").
				Str("name", deployment.Name).
				Str("namespace", deployment.Namespace).
				Msg("Deployment deleted")
		},
	})

	return informer
}
