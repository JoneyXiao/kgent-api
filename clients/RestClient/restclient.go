// We usually use RestClient in the case where aggregated API is used.
// This is a simple example of how to use the RestClient to list resources in a namespace.

// Use HTTP request to list/get/create/update/delete resources in a namespace.

// go run clients/RestClient/restclient.go --namespace=kube-system

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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
	namespace = flag.String("namespace", "default", "namespace to list pods from")
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Set up REST client configuration
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.APIPath = "/api"

	// Create REST client
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatalf("Error creating REST client: %v", err)
	}

	// Execute the REST request
	result := restClient.Get().
		Resource("pods").
		Namespace(*namespace).
		Do(ctx)

	if err := result.Error(); err != nil {
		log.Fatalf("Error executing REST request: %v", err)
	}

	// Parse the result
	podList := &corev1.PodList{}
	if err := result.Into(podList); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Display pod information
	if len(podList.Items) == 0 {
		fmt.Printf("No pods found in namespace %s\n", *namespace)
		return
	}

	fmt.Printf("Found %d pods in namespace %s:\n", len(podList.Items), *namespace)
	for _, pod := range podList.Items {
		printPodInfo(pod)
	}
}

func printPodInfo(pod corev1.Pod) {
	fmt.Printf("Pod: %s\n", pod.Name)
	fmt.Printf("  Status: %s\n", pod.Status.Phase)
	fmt.Printf("  Node: %s\n", pod.Spec.NodeName)
	fmt.Printf("  IP: %s\n", pod.Status.PodIP)

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

	// Print labels if any
	if len(pod.Labels) > 0 {
		fmt.Println("  Labels:")
		for k, v := range pod.Labels {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}

	fmt.Println()
}
