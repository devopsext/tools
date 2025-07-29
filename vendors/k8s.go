package vendors

import (
	"context"
	"time"

	"github.com/devopsext/tools/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sPodOptions struct {
	Namespace string
	Name      string
}

type K8sOptions struct {
	Config  string
	Timeout int
}

type K8s struct {
	options   K8sOptions
	logger    common.Logger
	clientset *kubernetes.Clientset
}

func newK8sClient(options K8sOptions) (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", options.Config)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func (k *K8s) CustomPodDelete(k8sOptions K8sOptions, k8sPodOptions K8sPodOptions) ([]byte, error) {

	clientset := k.clientset
	if clientset == nil || k8sOptions != k.options {
		cs, err := newK8sClient(k8sOptions)
		if err != nil {
			return nil, err
		}
		clientset = cs
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(k8sOptions.Timeout)*time.Second)
	defer cancel()

	opts := metav1.DeleteOptions{}

	rClient := clientset.CoreV1().RESTClient()
	r := rClient.Delete().
		UseProtobufAsDefaultIfPreferred(true).
		NamespaceIfScoped(k8sPodOptions.Namespace, true).
		Resource("pod").
		Name(k8sPodOptions.Name).
		Body(&opts).
		Do(ctx)

	return r.Raw()
}

func (k *K8s) PodDelete(options K8sPodOptions) ([]byte, error) {
	return k.CustomPodDelete(k.options, options)
}

func NewK8s(options K8sOptions, logger common.Logger) *K8s {

	clientset, _ := newK8sClient(options)

	return &K8s{
		options:   options,
		logger:    logger,
		clientset: clientset,
	}
}
