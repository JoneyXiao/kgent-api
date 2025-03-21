// DiscoveryClient is used to discover API resources available in the cluster.
// This client helps with querying the server for available API groups, versions, and resources.
// It's commonly used for building dynamic clients that need to interact with resources
// without prior knowledge of their API specifications.

// Use it together with RestMapper to GVR and GVK conversions.

// `kubectl api-resources` is a command that uses DiscoveryClient to list all API resources in the cluster.

// go run clients/DiscoveryClient/discoveryclient.go --resources

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Set up command line flags
	var kubeconfig *string
	var showResources *bool

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	showResources = flag.Bool("resources", false, "show API resources in addition to groups")
	flag.Parse()

	// Create a context with timeout
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Fatalf("Error creating discovery client: %v", err)
	}

	// Get server API groups
	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		log.Fatalf("Error fetching API groups: %v", err)
	}

	fmt.Println("=== API Groups ===")
	for _, group := range apiGroups.Groups {
		fmt.Printf("API Group: %s\n", group.Name)
		for _, version := range group.Versions {
			fmt.Printf("  Version: %s\n", version.GroupVersion)

			// Get resources for this API group version if requested
			if *showResources {
				resources, err := discoveryClient.ServerResourcesForGroupVersion(version.GroupVersion)
				if err != nil {
					fmt.Printf("    Error fetching resources: %v\n", err)
					continue
				}

				for _, resource := range resources.APIResources {
					namespaced := "cluster-scoped"
					if resource.Namespaced {
						namespaced = "namespaced"
					}

					verbs := ""
					for i, verb := range resource.Verbs {
						if i > 0 {
							verbs += ", "
						}
						verbs += string(verb)
					}

					fmt.Printf("    Resource: %s (%s)\n", resource.Name, namespaced)
					fmt.Printf("      Kind: %s\n", resource.Kind)
					fmt.Printf("      Verbs: %s\n", verbs)
				}
			}
		}
		fmt.Println()
	}

	// Get server version
	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		log.Fatalf("Error fetching server version: %v", err)
	}

	fmt.Println("=== Server Version ===")
	fmt.Printf("Version: %s\n", serverVersion.GitVersion)
	fmt.Printf("Platform: %s/%s\n", serverVersion.Platform, serverVersion.GoVersion)
	fmt.Printf("Build Date: %s\n", serverVersion.BuildDate)
}
