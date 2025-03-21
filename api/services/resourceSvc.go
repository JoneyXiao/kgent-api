package services

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
)

type ResourceService struct {
	restMapper *meta.RESTMapper
	client     *dynamic.DynamicClient
	fact       informers.SharedInformerFactory
}

func NewResourceService(restMapper *meta.RESTMapper, client *dynamic.DynamicClient, fact informers.SharedInformerFactory) *ResourceService {
	return &ResourceService{restMapper: restMapper, client: client, fact: fact}
}

func (r *ResourceService) ListResource(ctx context.Context, resourceOrKindArg string, ns string) ([]runtime.Object, error) {
	restMapping, err := r.mappingFor(resourceOrKindArg, r.restMapper)
	if err != nil {
		return nil, err
	}

	informer, err := r.fact.ForResource(restMapping.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get informer for resource %s: %w", resourceOrKindArg, err)
	}

	list, err := informer.Lister().ByNamespace(ns).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list %s resources: %w", resourceOrKindArg, err)
	}

	return list, nil
}

func (r *ResourceService) DeleteResource(ctx context.Context, resourceOrKindArg string, ns string, name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	ri, err := r.getResourceInterface(resourceOrKindArg, ns, r.client, r.restMapper)
	if err != nil {
		return err
	}

	err = ri.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete %s/%s: %w", resourceOrKindArg, name, err)
	}
	return nil
}

func (r *ResourceService) CreateResource(ctx context.Context, resourceOrKindArg string, yaml string) error {
	if yaml == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}

	obj := &unstructured.Unstructured{}
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(yaml), nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	ri, err := r.getResourceInterface(resourceOrKindArg, obj.GetNamespace(), r.client, r.restMapper)
	if err != nil {
		return err
	}

	_, err = ri.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", resourceOrKindArg, err)
	}
	return nil
}

func (r *ResourceService) GetGVR(resourceOrKindArg string) (*schema.GroupVersionResource, error) {
	if resourceOrKindArg == "" {
		return nil, fmt.Errorf("resource type cannot be empty")
	}

	restMapping, err := r.mappingFor(resourceOrKindArg, r.restMapper)
	if err != nil {
		return nil, err
	}

	return &restMapping.Resource, nil
}

// getResourceInterface returns the appropriate dynamic resource interface based on the resource type and namespace
func (r *ResourceService) getResourceInterface(resourceOrKindArg string, ns string, client dynamic.Interface, restMapper *meta.RESTMapper) (dynamic.ResourceInterface, error) {
	var ri dynamic.ResourceInterface

	restMapping, err := r.mappingFor(resourceOrKindArg, restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get RESTMapping for %s: %w", resourceOrKindArg, err)
	}

	// Determine if resource is namespaced or cluster-scoped
	if restMapping.Scope.Name() == "namespace" {
		ri = client.Resource(restMapping.Resource).Namespace(ns)
	} else {
		ri = client.Resource(restMapping.Resource)
	}

	return ri, nil
}

// mappingFor finds the REST mapping for a resource
func (r *ResourceService) mappingFor(resourceOrKindArg string, restMapper *meta.RESTMapper) (*meta.RESTMapping, error) {
	if resourceOrKindArg == "" {
		return nil, fmt.Errorf("resource type cannot be empty")
	}

	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)
	gvk := schema.GroupVersionKind{}

	if fullySpecifiedGVR != nil {
		gvk, _ = (*restMapper).KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		gvk, _ = (*restMapper).KindFor(groupResource.WithVersion(""))
	}
	if !gvk.Empty() {
		return (*restMapper).RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(resourceOrKindArg)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := (*restMapper).RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := (*restMapper).RESTMapping(groupKind, gvk.Version)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}
		return nil, err
	}

	return mapping, nil
}
