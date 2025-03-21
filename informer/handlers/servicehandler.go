package handlers

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// ServiceHandler implements ResourceEventHandler for Service resources
type ServiceHandler struct {
	Caller string
}

// OnAdd is called when a Service is added
func (h *ServiceHandler) OnAdd(obj interface{}, isInInitialList bool) {
	svc, ok := obj.(*v1.Service)
	if !ok {
		fmt.Println("Error: OnAdd received non-Service object")
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

	var serviceType string
	if len(string(svc.Spec.Type)) > 0 {
		serviceType = string(svc.Spec.Type)
	} else {
		serviceType = "ClusterIP" // Default type
	}

	fmt.Printf("[Caller: %s] [ServiceHandler] Service Added%s: %s/%s (type: %s)\n",
		caller,
		initialListMsg,
		svc.Namespace,
		svc.Name,
		serviceType)

	// Display service ports
	if len(svc.Spec.Ports) > 0 {
		fmt.Printf("  Ports: ")
		for i, port := range svc.Spec.Ports {
			if i > 0 {
				fmt.Printf(", ")
			}
			if len(port.Name) > 0 {
				fmt.Printf("%s:", port.Name)
			}
			fmt.Printf("%d->", port.Port)
			if port.TargetPort.IntVal != 0 {
				fmt.Printf("%d", port.TargetPort.IntVal)
			} else {
				fmt.Printf("%s", port.TargetPort.StrVal)
			}
			fmt.Printf("/%s", port.Protocol)
		}
		fmt.Println()
	}

	// Display selector if present
	if len(svc.Spec.Selector) > 0 {
		fmt.Println("  Selector:")
		for k, v := range svc.Spec.Selector {
			fmt.Printf("    %s: %s\n", k, v)
		}
	}
}

// OnUpdate is called when a Service is modified
func (h *ServiceHandler) OnUpdate(oldObj, newObj interface{}) {
	oldSvc, ok := oldObj.(*v1.Service)
	if !ok {
		fmt.Println("Error: OnUpdate received non-Service object for old object")
		return
	}

	newSvc, ok := newObj.(*v1.Service)
	if !ok {
		fmt.Println("Error: OnUpdate received non-Service object for new object")
		return
	}

	if newSvc.ResourceVersion == oldSvc.ResourceVersion {
		// No actual change, skip
		return
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [ServiceHandler] Service Updated: %s/%s\n",
		caller,
		newSvc.Namespace,
		newSvc.Name)

	// Check for type change
	if oldSvc.Spec.Type != newSvc.Spec.Type {
		fmt.Printf("  Service type changed: %s -> %s\n",
			oldSvc.Spec.Type,
			newSvc.Spec.Type)
	}

	// Check for ClusterIP change (service re-creation often happens)
	if oldSvc.Spec.ClusterIP != newSvc.Spec.ClusterIP {
		fmt.Printf("  ClusterIP changed: %s -> %s\n",
			oldSvc.Spec.ClusterIP,
			newSvc.Spec.ClusterIP)
	}
}

// OnDelete is called when a Service is deleted
func (h *ServiceHandler) OnDelete(obj interface{}) {
	svc, ok := obj.(*v1.Service)
	if !ok {
		// When a delete is observed, the object might be a DeletedFinalStateUnknown
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			fmt.Println("Error: OnDelete received non-Service and non-DeletedFinalStateUnknown object")
			return
		}

		svc, ok = tombstone.Obj.(*v1.Service)
		if !ok {
			fmt.Println("Error: DeletedFinalStateUnknown contained non-Service object")
			return
		}
	}

	caller := h.Caller
	if caller == "" {
		caller = "unknown"
	}

	fmt.Printf("[Caller: %s] [ServiceHandler] Service Deleted: %s/%s\n",
		caller,
		svc.Namespace,
		svc.Name)
}
