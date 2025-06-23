/*
Copyright © 2025 Danylo Liaudanskyi liaudanskyidanylo@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	namespace  string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Kubernetes deployments",
	Long: `List all Kubernetes deployments in the specified namespace.
	
Examples:
  # List deployments in default namespace
  k8s-controller-cli list
  
  # List deployments in specific namespace
  k8s-controller-cli list --namespace kube-system
  
  # Use custom kubeconfig file
  k8s-controller-cli list --kubeconfig /path/to/kubeconfig`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listDeployments(); err != nil {
			log.Error().Err(err).Msg("Failed to list deployments")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add flags
	listCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (default: $HOME/.kube/config)")
	listCmd.Flags().StringVar(&namespace, "namespace", "default", "Namespace to list deployments from")
}

// listDeployments lists all deployments in the specified namespace
func listDeployments() error {
	// Build kubeconfig path
	if kubeconfig == "" {
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE") // Windows
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file not found at %s", kubeconfig)
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// List deployments
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments in namespace '%s': %w", namespace, err)
	}

	// Print results
	fmt.Printf("Deployments in namespace '%s':\n", namespace)
	fmt.Println("NAME\t\tREADY\tUP-TO-DATE\tAVAILABLE\tAGE")

	for _, deployment := range deployments.Items {
		// Handle nil replicas
		var replicas int32
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}

		ready := fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, replicas)
		upToDate := fmt.Sprintf("%d", deployment.Status.UpdatedReplicas)
		available := fmt.Sprintf("%d", deployment.Status.AvailableReplicas)

		// Calculate age
		age := "Unknown"
		if !deployment.CreationTimestamp.IsZero() {
			duration := time.Since(deployment.CreationTimestamp.Time)
			if duration < time.Minute {
				age = fmt.Sprintf("%ds", int(duration.Seconds()))
			} else if duration < time.Hour {
				age = fmt.Sprintf("%dm", int(duration.Minutes()))
			} else if duration < 24*time.Hour {
				age = fmt.Sprintf("%dh", int(duration.Hours()))
			} else {
				age = fmt.Sprintf("%dd", int(duration.Hours()/24))
			}
		}

		fmt.Printf("%s\t\t%s\t%s\t\t%s\t\t%s\n",
			deployment.Name, ready, upToDate, available, age)
	}

	if len(deployments.Items) == 0 {
		fmt.Println("No deployments found.")
	}

	return nil
}
