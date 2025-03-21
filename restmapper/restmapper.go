// RestMapper is a client that helps with mapping between GroupVersionKind (GVK) and GroupVersionResource (GVR).
// It's commonly used for building dynamic clients that need to interact with resources
// without prior knowledge of their API specifications.
// This is a simple example of how to use the RestMapper to get the GVR, GVK and scope from the resource or kind argument.
// Once has the GVR, can use the dynamic client to interact with the resource.

// Use it together with DiscoveryClient to handle GVR and GVK conversions.

// go run restmapper/restmapper.go --namespace=kube-system --resource=pods

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Set up command line flags
	var kubeconfig *string
	var namespace *string
	var resourceArg *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	namespace = flag.String("namespace", "default", "namespace to list resources from")
	resourceArg = flag.String("resource", "pods", "resource type or kind to list (e.g. pods, deployments.apps, Pod, Deployment)")
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create clientset for discovery
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create dynamic client for resource access
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamic client: %v", err)
	}

	// Initialize REST mapper
	restMapper := InitRestMapper(clientset)

	// Get REST mapping for the requested resource
	restMapping, err := mappingFor(*resourceArg, &restMapper)
	if err != nil {
		log.Fatalf("Error getting REST mapping for %s: %v", *resourceArg, err)
	}

	if restMapping == nil {
		log.Fatalf("Could not find REST mapping for resource: %s", *resourceArg)
	}

	fmt.Printf("Resource Mapping Information:\n")
	fmt.Printf("  GVR: %s\n", restMapping.Resource)
	fmt.Printf("  GVK: %s\n", restMapping.GroupVersionKind)
	fmt.Printf("  Scope: %s\n\n", restMapping.Scope.Name())

	// Create a resource interface
	var resourceInterface dynamic.ResourceInterface
	if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resourceInterface = dynamicClient.Resource(restMapping.Resource).Namespace(*namespace)
		fmt.Printf("Listing %s in namespace %s:\n", restMapping.Resource.Resource, *namespace)
	} else {
		resourceInterface = dynamicClient.Resource(restMapping.Resource)
		fmt.Printf("Listing cluster-scoped %s:\n", restMapping.Resource.Resource)
	}

	// List resources
	resources, err := resourceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error listing resources: %v", err)
	}

	if len(resources.Items) == 0 {
		fmt.Println("No resources found.")
		return
	}

	// Display resources
	for _, item := range resources.Items {
		fmt.Printf("- %s\n", item.GetName())

		// Display more information if available
		if item.GetNamespace() != "" {
			fmt.Printf("  Namespace: %s\n", item.GetNamespace())
		}

		// Display labels if any
		if labels := item.GetLabels(); len(labels) > 0 {
			fmt.Println("  Labels:")
			for k, v := range labels {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
	}
}

// InitRestMapper initializes a REST mapper from discovery client
func InitRestMapper(clientSet *kubernetes.Clientset) meta.RESTMapper {
	gr, err := restmapper.GetAPIGroupResources(clientSet.Discovery())
	if err != nil {
		log.Fatalf("Error getting API group resources: %v", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return mapper
}

// mappingFor gets the REST mapping for a resource or kind argument
func mappingFor(resourceOrKindArg string, restMapper *meta.RESTMapper) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)
	gvk := schema.GroupVersionKind{}

	if fullySpecifiedGVR != nil {
		var err error
		gvk, err = (*restMapper).KindFor(*fullySpecifiedGVR)
		if err != nil {
			fmt.Printf("Warning: Could not get kind for GVR %s: %v\n", fullySpecifiedGVR, err)
		}
	}

	if gvk.Empty() {
		var err error
		gvk, err = (*restMapper).KindFor(groupResource.WithVersion(""))
		if err != nil {
			fmt.Printf("Warning: Could not get kind for group resource %s: %v\n", groupResource, err)
		}
	}

	if !gvk.Empty() {
		return (*restMapper).RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	// Try a direct mapping as fallback
	fmt.Printf("Trying direct mapping for group: %s, resource: %s\n", groupResource.Group, groupResource.Resource)
	return (*restMapper).RESTMapping(schema.GroupKind{
		Group: groupResource.Group,
		Kind:  groupResource.Resource,
	}, "")
}
