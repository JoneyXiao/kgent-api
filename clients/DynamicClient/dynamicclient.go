// DynamicClient is a commonly used client that can be used to interact with any resource in the cluster.
// This is a simple example of how to use the DynamicClient to list resources in a namespace.

// GVR: Group Version Resource
// For pod, the GVR is core/v1/pods
// For deployment, the GVR is apps/v1/deployments

// go run clients/DynamicClient/dynamicclient.go --namespace=kube-system --group=apps --version=v1 --resource=deployments

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Set up command line flags
	var kubeconfig *string
	var namespace, group, version, resource *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	namespace = flag.String("namespace", "kube-system", "namespace to list resources from")
	group = flag.String("group", "apps", "API group of the resource")
	version = flag.String("version", "v1", "API version of the resource")
	resource = flag.String("resource", "deployments", "resource type to list")

	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	// Define the GroupVersionResource
	gvr := schema.GroupVersionResource{
		Group:    *group,
		Version:  *version,
		Resource: *resource,
	}

	// Get resources from the specified namespace
	resources, err := dynamicClient.Resource(gvr).Namespace(*namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing resources: %v", err)
	}

	if len(resources.Items) == 0 {
		fmt.Printf("No %s resources found in namespace %s\n", *resource, *namespace)
		return
	}

	// Display resource information
	fmt.Printf("Found %d %s resources in namespace %s:\n", len(resources.Items), *resource, *namespace)
	for _, item := range resources.Items {
		fmt.Printf("Resource: %s\n", item.GetName())
		fmt.Printf("  UID: %s\n", item.GetUID())
		fmt.Printf("  Created: %s\n", item.GetCreationTimestamp())

		// Print labels if any
		if labels := item.GetLabels(); len(labels) > 0 {
			fmt.Println("  Labels:")
			for k, v := range labels {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}

		// Print annotations if any
		if annotations := item.GetAnnotations(); len(annotations) > 0 {
			fmt.Println("  Annotations:")
			for k, v := range annotations {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}

		fmt.Println()
	}
}
