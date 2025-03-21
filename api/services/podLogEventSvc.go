package services

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodLogEventService struct {
	client *kubernetes.Clientset
}

func NewPodLogEventService(client *kubernetes.Clientset) *PodLogEventService {
	return &PodLogEventService{client: client}
}

func (p *PodLogEventService) GetLogs(ctx context.Context, ns, podname, container string, tailLine int64) (*rest.Request, error) {
	if podname == "" {
		return nil, fmt.Errorf("pod name cannot be empty")
	}

	// If container is empty, don't specify it in options to get logs from default container
	options := &v1.PodLogOptions{Follow: false, TailLines: &tailLine}
	if container != "" {
		options.Container = container
	}

	req := p.client.CoreV1().Pods(ns).GetLogs(podname, options)
	return req, nil
}

func (p *PodLogEventService) GetEvents(ctx context.Context, ns, podname string) ([]string, error) {
	if podname == "" {
		return nil, fmt.Errorf("pod name cannot be empty")
	}

	events, err := p.client.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podname),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var podEvents []string
	for _, event := range events.Items {
		if event.Type == "Warning" {
			podEvents = append(podEvents, event.Message)
		}
	}

	return podEvents, nil
}
