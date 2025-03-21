package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sConfig struct {
	*rest.Config
	*kubernetes.Clientset
	*dynamic.DynamicClient
	meta.RESTMapper
	informers.SharedInformerFactory
	e error
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

// InitRestConfig initializes Kubernetes REST config
func (k *K8sConfig) InitRestConfig(optfuncs ...K8sConfigOptionFunc) *K8sConfig {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		k.e = errors.Wrap(err, "failed to build config from flags")
		return k
	}

	k.Config = config
	for _, optfunc := range optfuncs {
		optfunc(k)
	}
	return k
}

// InitConfigInCluster initializes in-cluster configuration
func (k *K8sConfig) InitConfigInCluster() *K8sConfig {
	config, err := rest.InClusterConfig()
	if err != nil {
		k.e = errors.Wrap(err, "failed to get in-cluster config")
		return k
	}
	k.Config = config
	return k
}

func (k *K8sConfig) Error() error {
	return k.e
}

// InitClientSet initializes Kubernetes clientset
func (k *K8sConfig) InitClientSet() *kubernetes.Clientset {
	if k.Config == nil {
		k.e = errors.New("k8s config is nil")
		return nil
	}

	clientSet, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		k.e = errors.Wrap(err, "failed to create clientset")
		return nil
	}
	k.Clientset = clientSet
	return clientSet
}

// InitDynamicClient initializes dynamic client
func (k *K8sConfig) InitDynamicClient() *dynamic.DynamicClient {
	if k.Config == nil {
		k.e = errors.New("k8s config is nil")
		return nil
	}

	dynamicClient, err := dynamic.NewForConfig(k.Config)
	if err != nil {
		k.e = errors.Wrap(err, "failed to create dynamic client")
		return nil
	}
	k.DynamicClient = dynamicClient
	return dynamicClient
}

// InitRestMapper initializes REST mapper for API resources
func (k *K8sConfig) InitRestMapper() meta.RESTMapper {
	if k.Clientset == nil {
		k.InitClientSet()
		if k.e != nil {
			return nil
		}
	}

	gr, err := restmapper.GetAPIGroupResources(k.Clientset.Discovery())
	if err != nil {
		k.e = errors.Wrap(err, "failed to get API group resources")
		return nil
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	k.RESTMapper = mapper
	return mapper
}

// InitInformer initializes shared informer factory
func (k *K8sConfig) InitInformer() informers.SharedInformerFactory {
	if k.Clientset == nil {
		k.InitClientSet()
		if k.e != nil {
			return nil
		}
	}

	fact := informers.NewSharedInformerFactory(k.Clientset, 0)

	// Initialize default informers as needed
	fact.Core().V1().Pods().Informer()
	fact.Core().V1().Services().Informer()
	fact.Apps().V1().Deployments().Informer()

	ch := make(chan struct{})
	fact.Start(ch)
	fact.WaitForCacheSync(ch)

	k.SharedInformerFactory = fact
	return fact
}

type K8sConfigOptionFunc func(k *K8sConfig)

func WithQps(qps float32) K8sConfigOptionFunc {
	return func(k *K8sConfig) {
		if k.Config != nil {
			k.QPS = qps
		}
	}
}

func WithBurst(b int) K8sConfigOptionFunc {
	return func(k *K8sConfig) {
		if k.Config != nil {
			k.Burst = b
		}
	}
}

func WithTimeout(timeout int) K8sConfigOptionFunc {
	return func(k *K8sConfig) {
		if k.Config != nil {
			k.Timeout = time.Duration(timeout) * time.Second
		}
	}
}
