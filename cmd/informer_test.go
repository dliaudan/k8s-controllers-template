package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func TestInformerCommandDefined(t *testing.T) {
	if InformerCmd == nil {
		t.Fatal("InformerCmd should be defined")
	}
	if InformerCmd.Use != "informer" {
		t.Errorf("expected command use 'informer', got %s", InformerCmd.Use)
	}

	// Test flags
	inClusterFlag := InformerCmd.Flags().Lookup("in-cluster")
	if inClusterFlag == nil {
		t.Error("expected 'in-cluster' flag to be defined")
	}

	kubeconfigFlag := InformerCmd.Flags().Lookup("kubeconfig")
	if kubeconfigFlag == nil {
		t.Error("expected 'kubeconfig' flag to be defined")
	}

	namespaceFlag := InformerCmd.Flags().Lookup("namespace")
	if namespaceFlag == nil {
		t.Error("expected 'namespace' flag to be defined")
	}

	resyncPeriodFlag := InformerCmd.Flags().Lookup("resync-period")
	if resyncPeriodFlag == nil {
		t.Error("expected 'resync-period' flag to be defined")
	}
}

func TestInformerCommandHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run help command
	InformerCmd.SetArgs([]string{"--help"})
	InformerCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("failed to read from pipe: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got empty string")
	}
}

func TestInformerWithEnvTest(t *testing.T) {
	// Setup envtest
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false, // Set to false since we don't have CRDs
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer testEnv.Stop()

	// Create clientset
	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// Set test namespace
	namespaceInformer = "default"
	resyncPeriodInformer = 1 * time.Second

	// Create informer
	informer := createDeploymentInformer(clientset)
	require.NotNil(t, informer)

	// Track events
	var events []string
	var eventMutex sync.Mutex

	// Add custom event handler to track events
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "ADDED:"+deployment.Name)
			eventMutex.Unlock()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment := newObj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "UPDATED:"+deployment.Name)
			eventMutex.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "DELETED:"+deployment.Name)
			eventMutex.Unlock()
		},
	})

	// Start informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	go informer.Run(stopCh)

	// Wait for informer to sync
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		t.Fatal("Failed to sync informer cache")
	}

	// Test deployment operations
	testDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envtest-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "envtest"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "envtest"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}

	// Create deployment
	_, err = clientset.AppsV1().Deployments("default").Create(context.Background(), testDeployment, metav1.CreateOptions{})
	require.NoError(t, err)

	// Wait for ADDED event
	err = wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
		eventMutex.Lock()
		defer eventMutex.Unlock()
		for _, event := range events {
			if event == "ADDED:envtest-deployment" {
				return true, nil
			}
		}
		return false, nil
	})
	assert.NoError(t, err, "Expected ADDED event for envtest-deployment")

	// Verify deployment exists
	deployment, err := clientset.AppsV1().Deployments("default").Get(context.Background(), "envtest-deployment", metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, "envtest-deployment", deployment.Name)

	// Update deployment
	deployment.Spec.Replicas = int32Ptr(3)
	_, err = clientset.AppsV1().Deployments("default").Update(context.Background(), deployment, metav1.UpdateOptions{})
	require.NoError(t, err)

	// Wait for UPDATED event
	err = wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
		eventMutex.Lock()
		defer eventMutex.Unlock()
		for _, event := range events {
			if event == "UPDATED:envtest-deployment" {
				return true, nil
			}
		}
		return false, nil
	})
	assert.NoError(t, err, "Expected UPDATED event for envtest-deployment")

	// Delete deployment
	err = clientset.AppsV1().Deployments("default").Delete(context.Background(), "envtest-deployment", metav1.DeleteOptions{})
	require.NoError(t, err)

	// Wait for DELETED event
	err = wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
		eventMutex.Lock()
		defer eventMutex.Unlock()
		for _, event := range events {
			if event == "DELETED:envtest-deployment" {
				return true, nil
			}
		}
		return false, nil
	})
	assert.NoError(t, err, "Expected DELETED event for envtest-deployment")

	// Verify all events were received
	eventMutex.Lock()
	defer eventMutex.Unlock()
	assert.Contains(t, events, "ADDED:envtest-deployment")
	assert.Contains(t, events, "UPDATED:envtest-deployment")
	assert.Contains(t, events, "DELETED:envtest-deployment")
}

func TestInformerWithEnvTestAllNamespaces(t *testing.T) {
	// Setup envtest
	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	require.NoError(t, err)
	defer testEnv.Stop()

	// Create clientset
	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	// Create test namespace
	testNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), testNamespace, metav1.CreateOptions{})
	require.NoError(t, err)

	// Set test configuration for all namespaces
	namespaceInformer = "" // Empty means all namespaces
	resyncPeriodInformer = 1 * time.Second

	// Create informer
	informer := createDeploymentInformer(clientset)
	require.NotNil(t, informer)

	// Track events
	var events []string
	var eventMutex sync.Mutex

	// Add custom event handler to track events
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "ADDED:"+deployment.Namespace+":"+deployment.Name)
			eventMutex.Unlock()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			deployment := newObj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "UPDATED:"+deployment.Namespace+":"+deployment.Name)
			eventMutex.Unlock()
		},
		DeleteFunc: func(obj interface{}) {
			deployment := obj.(*appsv1.Deployment)
			eventMutex.Lock()
			events = append(events, "DELETED:"+deployment.Namespace+":"+deployment.Name)
			eventMutex.Unlock()
		},
	})

	// Start informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	go informer.Run(stopCh)

	// Wait for informer to sync
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		t.Fatal("Failed to sync informer cache")
	}

	// Create deployments in different namespaces
	deployments := []struct {
		name      string
		namespace string
	}{
		{"deployment-1", "default"},
		{"deployment-2", "test-namespace"},
	}

	for _, dep := range deployments {
		testDeployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dep.name,
				Namespace: dep.namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": dep.name},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": dep.name},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "nginx",
								Image: "nginx:latest",
							},
						},
					},
				},
			},
		}

		_, err = clientset.AppsV1().Deployments(dep.namespace).Create(context.Background(), testDeployment, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	// Wait for ADDED events
	for _, dep := range deployments {
		err = wait.PollImmediate(100*time.Millisecond, 5*time.Second, func() (bool, error) {
			eventMutex.Lock()
			defer eventMutex.Unlock()
			expectedEvent := "ADDED:" + dep.namespace + ":" + dep.name
			for _, event := range events {
				if event == expectedEvent {
					return true, nil
				}
			}
			return false, nil
		})
		assert.NoError(t, err, "Expected ADDED event for %s in namespace %s", dep.name, dep.namespace)
	}

	// Verify all events were received
	eventMutex.Lock()
	defer eventMutex.Unlock()
	assert.Contains(t, events, "ADDED:default:deployment-1")
	assert.Contains(t, events, "ADDED:test-namespace:deployment-2")

	// Cleanup
	for _, dep := range deployments {
		err = clientset.AppsV1().Deployments(dep.namespace).Delete(context.Background(), dep.name, metav1.DeleteOptions{})
		require.NoError(t, err)
	}
}

func TestInformerCommandIntegration(t *testing.T) {
	// Test that the command is properly added to root
	rootCmd.AddCommand(InformerCmd)

	// Verify informer command is in root command's children
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "informer" {
			found = true
			break
		}
	}

	if !found {
		t.Error("informer command should be added to root command")
	}
}

// Helper function for int32 pointers
func int32Ptr(i int32) *int32 {
	return &i
}
