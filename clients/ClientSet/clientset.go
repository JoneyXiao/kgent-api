// ClientSet is the most common way to interact with the Kubernetes core resources like pods, services, deployments, etc.
// This is a simple example of how to use the ClientSet to list pods in a namespace.

// GVK: Group Version Kind
// For pod, the GVK is core/v1/Pod = clientset.CoreV1().Pods()
// For deployment, the GVK is apps/v1/Deployment = clientset.AppsV1().Deployments()

// go run clients/ClientSet/clientset.go --namespace=kube-system

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Set up command line flags
	var kubeconfig *string
	var namespace *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace = flag.String("namespace", "kube-system", "namespace to list pods from")
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Get pods from the specified namespace
	pods, err := clientset.CoreV1().Pods(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
	}

	if len(pods.Items) == 0 {
		fmt.Printf("No pods found in namespace %s\n", *namespace)
		return
	}

	// Display pod information
	fmt.Printf("Found %d pods in namespace %s:\n", len(pods.Items), *namespace)
	for _, pod := range pods.Items {
		printPodInfo(pod)
	}
}

func printPodInfo(pod corev1.Pod) {
	fmt.Printf("Pod: %s\n", pod.Name)
	fmt.Printf("  Status: %s\n", pod.Status.Phase)
	fmt.Printf("  Node: %s\n", pod.Spec.NodeName)

	if len(pod.Status.ContainerStatuses) > 0 {
		fmt.Printf("  Containers: %d\n", len(pod.Status.ContainerStatuses))
		for _, container := range pod.Status.ContainerStatuses {
			ready := "Not Ready"
			if container.Ready {
				ready = "Ready"
			}
			fmt.Printf("    - %s: %s (Restarts: %d)\n",
				container.Name,
				ready,
				container.RestartCount)
		}
	}
	fmt.Println()
}
