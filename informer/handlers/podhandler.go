package handlers

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// PodHandler implements ResourceEventHandler for Pod resources
type PodHandler struct {
	Caller string
}

// OnAdd is called when a Pod is added
func (h *PodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		fmt.Println("Error: OnAdd received non-Pod object")
		return
	}

	initialListMsg := ""
	if isInInitialList {
		initialListMsg = " (initial list)"
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [PodHandler] Pod Added%s: %s/%s (phase: %s, containers: %d)\n",
		caller,
		initialListMsg,
		pod.Namespace,
		pod.Name,
		pod.Status.Phase,
		len(pod.Spec.Containers))
}

// OnUpdate is called when a Pod is modified
func (h *PodHandler) OnUpdate(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		fmt.Println("Error: OnUpdate received non-Pod object for old object")
		return
	}

	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		fmt.Println("Error: OnUpdate received non-Pod object for new object")
		return
	}

	if newPod.ResourceVersion == oldPod.ResourceVersion {
		// No actual change, skip
		return
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	// Report meaningful changes
	if oldPod.Status.Phase != newPod.Status.Phase {
		fmt.Printf("[Caller: %s] [PodHandler] Pod Phase Changed: %s/%s (%s -> %s)\n",
			caller,
			newPod.Namespace,
			newPod.Name,
			oldPod.Status.Phase,
			newPod.Status.Phase)
	} else {
		fmt.Printf("[Caller: %s] [PodHandler] Pod Updated: %s/%s (rv: %s)\n",
			caller,
			newPod.Namespace,
			newPod.Name,
			newPod.ResourceVersion)
	}
}

// OnDelete is called when a Pod is deleted
func (h *PodHandler) OnDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		// When a delete is observed, the object might be a DeletedFinalStateUnknown
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			fmt.Println("Error: OnDelete received non-Pod and non-DeletedFinalStateUnknown object")
			return
		}

		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			fmt.Println("Error: DeletedFinalStateUnknown contained non-Pod object")
			return
		}
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [PodHandler] Pod Deleted: %s/%s\n",
		caller,
		pod.Namespace,
		pod.Name)
}

// NewPodHandler is an alternative handler implementation for pods
type NewPodHandler struct {
	Caller string
}

// OnAdd is called when a Pod is added
func (h *NewPodHandler) OnAdd(obj interface{}, isInInitialList bool) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		fmt.Println("Error: NewPodHandler OnAdd received non-Pod object")
		return
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [NewPodHandler] Pod Added: %s/%s (created: %s)\n",
		caller,
		pod.Namespace,
		pod.Name,
		pod.CreationTimestamp.Format(time.RFC3339))

	// Show container information
	if len(pod.Spec.Containers) > 0 {
		fmt.Printf("  Containers: ")
		for i, container := range pod.Spec.Containers {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s (image: %s)", container.Name, container.Image)
		}
		fmt.Println()
	}
}

// OnUpdate is called when a Pod is modified
func (h *NewPodHandler) OnUpdate(oldObj, newObj interface{}) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [NewPodHandler] Pod Updated: %s/%s\n",
		caller,
		newPod.Namespace,
		newPod.Name)
}

// OnDelete is called when a Pod is deleted
func (h *NewPodHandler) OnDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		// Check for tombstone
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		pod, ok = tombstone.Obj.(*v1.Pod)
		if !ok {
			return
		}
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [NewPodHandler] Pod Deleted: %s/%s\n",
		caller,
		pod.Namespace,
		pod.Name)
}
