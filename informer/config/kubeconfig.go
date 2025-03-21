package config

import (
	"flag"
	"path/filepath"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sConfig struct {
	KubeConfigPath string
	Namespace      string
	Config         *rest.Config
	Clientset      *kubernetes.Clientset
	err            error
}

func NewK8sConfig() *K8sConfig {
	var kubeconfig *string
	var namespace *string

	// Set up default kubeconfig path
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	// Set up default namespace
	namespace = flag.String("namespace", "default", "namespace to watch resources from")

	// Parse flags if they haven't been parsed yet
	if !flag.Parsed() {
		flag.Parse()
	}

	return &K8sConfig{
		KubeConfigPath: *kubeconfig,
		Namespace:      *namespace,
	}
}

// Initialize Kubernetes REST config
func (k *K8sConfig) InitRestConfig() *K8sConfig {
	var err error
	var config *rest.Config

	// Try to use in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", k.KubeConfigPath)
		if err != nil {
			k.err = errors.Wrap(err, "failed to build kubeconfig")
			return k
		}
	}

	k.Config = config
	return k
}

// Return any error that occurred during configuration
func (k *K8sConfig) Error() error {
	return k.err
}

// Initialize Kubernetes clientset
func (k *K8sConfig) InitClientSet() *kubernetes.Clientset {
	if k.Config == nil {
		k.err = errors.New("kubernetes config is nil, call InitRestConfig first")
		return nil
	}

	clientset, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		k.err = errors.Wrap(err, "failed to create kubernetes clientset")
		return nil
	}

	k.Clientset = clientset
	return clientset
}
