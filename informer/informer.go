// This example demonstrates different ways to use informers in Kubernetes client-go.
// Informers provide a way to watch for changes to Kubernetes resources
// and react to those changes with event handlers.

// go run informer.go --type=all --namespace=default

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kgent-api/informer/config"
	"kgent-api/informer/handlers"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// basicInformer demonstrates the simplest informer setup with a single handler
// It uses the low-level cache.NewInformerWithOptions API
func basicInformer(lw *cache.ListWatch, stopCh <-chan struct{}) {
	fmt.Println("Running basic informer example...")

	// Configure the informer with our ListWatch and handler
	options := cache.InformerOptions{
		ListerWatcher: lw,                                            // Tells the informer what resources to watch
		ObjectType:    &v1.Pod{},                                     // Type of object to watch (Pod)
		ResyncPeriod:  time.Minute * 30,                              // How often to resync (full relist)
		Handler:       &handlers.PodHandler{Caller: "basicInformer"}, // Event handler
	}

	// Create a new informer with these options
	_, informer := cache.NewInformerWithOptions(options)

	// Run the informer
	go informer.Run(stopCh)

	// Wait for the informer to sync its cache
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		log.Fatal("Timed out waiting for caches to sync in basicInformer")
	}

	fmt.Println("Basic informer cache has synced and is running\n")
}

// sharedInformer demonstrates how to use a SharedInformer with multiple handlers
// SharedInformers allow multiple controllers to watch the same resources
// without duplicating API traffic or caching
func sharedInformer(lw *cache.ListWatch, stopCh <-chan struct{}) {
	fmt.Println("Running shared informer example...")

	// Create a shared informer that watches Pods
	sharedInformer := cache.NewSharedInformer(
		lw,             // Source of events
		&v1.Pod{},      // Type of objects to watch
		time.Minute*15, // How often to resync
	)

	// Add multiple event handlers to the same informer
	sharedInformer.AddEventHandler(&handlers.PodHandler{Caller: "sharedInformer"})
	sharedInformer.AddEventHandler(&handlers.NewPodHandler{Caller: "sharedInformer"})

	// Run the informer
	go sharedInformer.Run(stopCh)

	// Wait for the informer to sync its cache
	if !cache.WaitForCacheSync(stopCh, sharedInformer.HasSynced) {
		log.Fatal("Timed out waiting for caches to sync in sharedInformer")
	}

	fmt.Println("Shared informer cache has synced and is running\n")
}

// sharedInformerFactory demonstrates how to use a SharedInformerFactory
// The factory creates informers for multiple resource types
// and manages their lifecycle
func sharedInformerFactory(client *kubernetes.Clientset, namespace string, stopCh <-chan struct{}) {
	fmt.Println("Running shared informer factory example...")

	// Create a shared informer factory for the specified namespace
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,                             // Kubernetes client
		time.Minute*10,                     // Resync period
		informers.WithNamespace(namespace), // Only watch resources in this namespace
	)

	// Get informers for specific resource types from the factory
	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(&handlers.PodHandler{Caller: "sharedInformerFactory"})
	podInformer.Informer().AddEventHandler(&handlers.NewPodHandler{Caller: "sharedInformerFactory"})

	svcInformer := factory.Core().V1().Services()
	svcInformer.Informer().AddEventHandler(&handlers.ServiceHandler{Caller: "sharedInformerFactory"})

	// Start all informers in the factory
	factory.Start(stopCh)

	// Wait for caches to sync
	cachesSynced := []cache.InformerSynced{
		podInformer.Informer().HasSynced,
		svcInformer.Informer().HasSynced,
	}
	if !cache.WaitForCacheSync(stopCh, cachesSynced...) {
		log.Fatal("Timed out waiting for caches to sync in sharedInformerFactory")
		return
	}

	fmt.Println("Shared informer factory caches have synced and are running\n")
}

// sharedInformerFactoryLister demonstrates how to use Listers with SharedInformerFactory
// Listers provide a cached, indexed access to resources for better performance
func sharedInformerFactoryLister(client *kubernetes.Clientset, namespace string, stopCh <-chan struct{}) {
	fmt.Println("Running shared informer factory with lister example...")

	// Create a shared informer factory
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		time.Minute*10,
		informers.WithNamespace(namespace),
	)

	// Get pod informer from the factory
	podInformer := factory.Core().V1().Pods()
	podInformer.Informer().AddEventHandler(&handlers.PodHandler{Caller: "sharedInformerFactoryLister"})

	// Start informers
	factory.Start(stopCh)

	// Wait for caches to sync
	if !cache.WaitForCacheSync(stopCh, podInformer.Informer().HasSynced) {
		log.Fatal("Timed out waiting for caches to sync in sharedInformerFactoryLister")
		return
	}

	// Use the lister to get objects from the cache
	podList, err := podInformer.Lister().List(labels.Everything())
	if err != nil {
		log.Fatalf("Error listing pods: %v", err)
		return
	}

	fmt.Printf("Found %d pods in namespace %s through lister:\n", len(podList), namespace)
	for _, pod := range podList {
		fmt.Printf("- %s (status: %s)\n", pod.Name, pod.Status.Phase)
	}
	fmt.Println("\n")
}

// sharedInformerFactoryForResource demonstrates using generic ForResource
// method when informers for specific types are not available
// create a informer using GVR
func sharedInformerFactoryForResource(client *kubernetes.Clientset, namespace string, stopCh <-chan struct{}) {
	fmt.Println("Running shared informer factory for generic resource example...")

	// Create a shared informer factory
	factory := informers.NewSharedInformerFactoryWithOptions(
		client,
		time.Minute*10,
		informers.WithNamespace(namespace),
	)

	// Define the resource to watch using GroupVersionResource
	// For pods: "" (empty group), "v1", "pods"
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	// Get a generic informer for this resource
	informer, err := factory.ForResource(gvr)
	if err != nil {
		log.Fatalf("Error getting informer for %v: %v", gvr, err)
		return
	}

	// Add event handler with basic event functions
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Printf("[Generic handler] Pod added: %s\n", key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				fmt.Printf("[Generic handler] Pod updated: %s\n", key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				fmt.Printf("[Generic handler] Pod deleted: %s\n", key)
			}
		},
	})

	// Start the informers
	factory.Start(stopCh)

	// Wait for caches to sync
	if !cache.WaitForCacheSync(stopCh, informer.Informer().HasSynced) {
		log.Fatal("Timed out waiting for caches to sync in sharedInformerFactoryForResource")
		return
	}

	// Use the lister to get objects from the cache
	list, err := informer.Lister().List(labels.Everything())
	if err != nil {
		log.Fatalf("Error listing resources: %v", err)
		return
	}

	fmt.Printf("Found %d pods in namespace %s through generic lister:\n", len(list), namespace)
	for _, item := range list {
		fmt.Printf("- %s\n", item.(metav1.Object).GetName())
	}
	fmt.Println("\n")
}

func main() {
	// Parse command line flags
	exampleType := flag.String("type", "all",
		"Type of informer example to run: basic, shared, factory, lister, resource, all")
	flag.Parse()

	// Initialize Kubernetes client
	kubeConfig := config.NewK8sConfig()
	clientset := kubeConfig.InitRestConfig().InitClientSet()
	namespace := kubeConfig.Namespace

	// Check for configuration errors
	if err := kubeConfig.Error(); err != nil {
		log.Fatalf("Error initializing Kubernetes client: %v", err)
	}

	// Create a ListWatch for pods in the specified namespace
	lw := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"pods",
		namespace,
		fields.Everything(),
	)

	// Setup signal handling for graceful shutdown
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Run the requested informer example(s)
	switch *exampleType {
	case "basic":
		basicInformer(lw, stopCh)
	case "shared":
		sharedInformer(lw, stopCh)
	case "factory":
		sharedInformerFactory(clientset, namespace, stopCh)
	case "lister":
		sharedInformerFactoryLister(clientset, namespace, stopCh)
	case "resource":
		sharedInformerFactoryForResource(clientset, namespace, stopCh)
	case "all":
		basicInformer(lw, stopCh)
		sharedInformer(lw, stopCh)
		sharedInformerFactory(clientset, namespace, stopCh)
		sharedInformerFactoryLister(clientset, namespace, stopCh)
		sharedInformerFactoryForResource(clientset, namespace, stopCh)
	default:
		log.Fatalf("Unknown informer type: %s", *exampleType)
	}

	fmt.Println("\nInformers are running. Press Ctrl+C to stop...")

	// Wait for termination signal
	<-sigCh
	fmt.Println("\nReceived termination signal. Shutting down informers...")
	close(stopCh)

	// Allow time for graceful shutdown
	time.Sleep(2 * time.Second)
	fmt.Println("All informers stopped.")
}
