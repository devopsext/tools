package vendors

import (
	"context"
	"fmt"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sResourceOptions struct {
	Kind      string
	Namespace string
	Name      string
}

type K8sResourceDeleteOptions struct {
	K8sResourceOptions
}

type K8sResourceScaleOptions struct {
	K8sResourceOptions
	Replicas int
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

const (
	K8sResourcePods         = "pods"
	K8sResourceDeployments  = "deployments"
	K8sResourceReplicaSets  = "replicasets"
	K8sResourceStatefulSets = "statefulsets"
	K8sResourceDaemonSets   = "daemonsets"
)

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

func (k *K8s) getClientCtx(options K8sOptions) (*kubernetes.Clientset, context.Context, context.CancelFunc, error) {

	clientset := k.clientset
	if clientset == nil || options != k.options {
		cs, err := newK8sClient(options)
		if err != nil {
			return nil, nil, nil, err
		}
		clientset = cs
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(options.Timeout)*time.Second)
	return clientset, ctx, cancel, nil
}

func (k *K8s) CustomResourceDelete(options K8sOptions, deleteOptions K8sResourceDeleteOptions) ([]byte, error) {

	clientset, ctx, cancel, err := k.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	policy := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{
		PropagationPolicy: &policy,
	}

	var rClient rest.Interface

	switch deleteOptions.Kind {
	case K8sResourceDeployments:
		rClient = clientset.AppsV1().RESTClient()
	default:
		rClient = clientset.CoreV1().RESTClient()
	}

	r := rClient.Delete().
		UseProtobufAsDefaultIfPreferred(false).
		NamespaceIfScoped(deleteOptions.Namespace, !utils.IsEmpty(deleteOptions.Namespace)).
		Resource(deleteOptions.Kind).
		Name(deleteOptions.Name).
		Body(&opts).
		Do(ctx)

	return r.Raw()
}

func (k *K8s) ResourceDelete(options K8sResourceDeleteOptions) ([]byte, error) {
	return k.CustomResourceDelete(k.options, options)
}

func (k *K8s) CustomResourceScale(options K8sOptions, scaleOptions K8sResourceScaleOptions) ([]byte, error) {

	clientset, ctx, cancel, err := k.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	getOpts := metav1.GetOptions{}
	rClient := clientset.AppsV1().RESTClient()

	rGet := rClient.Get().
		UseProtobufAsDefaultIfPreferred(false).
		NamespaceIfScoped(scaleOptions.Namespace, !utils.IsEmpty(scaleOptions.Namespace)).
		Resource(scaleOptions.Kind).
		Name(scaleOptions.Name).
		VersionedParams(&getOpts, scheme.ParameterCodec).
		Do(ctx)

	err = rGet.Error()
	if err != nil {
		return rGet.Raw()
	}

	var currentReplicas int32
	desiredReplicasInt32 := int32(scaleOptions.Replicas)

	var obj interface{}

	switch scaleOptions.Kind {
	case K8sResourceDeployments:

		var d appsv1.Deployment
		err = rGet.Into(&d)
		if err != nil {
			return nil, err
		}
		currentReplicas = *d.Spec.Replicas
		d.Spec.Replicas = &desiredReplicasInt32
		obj = &d

	case K8sResourceReplicaSets:

		var rs appsv1.ReplicaSet
		err = rGet.Into(&rs)
		if err != nil {
			return nil, err
		}
		currentReplicas = *rs.Spec.Replicas
		rs.Spec.Replicas = &desiredReplicasInt32
		obj = &rs

	case K8sResourceStatefulSets:

		var ss appsv1.StatefulSet
		err = rGet.Into(&ss)
		if err != nil {
			return nil, err
		}

		currentReplicas = *ss.Spec.Replicas
		ss.Spec.Replicas = &desiredReplicasInt32
		obj = &ss

	default:
		return nil, fmt.Errorf("unsupported kind for scaling: %s", scaleOptions.Kind)
	}

	if currentReplicas == desiredReplicasInt32 {
		return nil, fmt.Errorf("current replicas are equal to desired replicas")
	}

	updateOpts := metav1.UpdateOptions{}
	rPut := rClient.Put().
		UseProtobufAsDefaultIfPreferred(false).
		NamespaceIfScoped(scaleOptions.Namespace, !utils.IsEmpty(scaleOptions.Namespace)).
		Resource(scaleOptions.Kind).
		Name(scaleOptions.Name).
		VersionedParams(&updateOpts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx)

	return rPut.Raw()
}

func (k *K8s) ResourceScale(options K8sResourceScaleOptions) ([]byte, error) {
	return k.CustomResourceScale(k.options, options)
}

func NewK8s(options K8sOptions, logger common.Logger) *K8s {

	clientset, _ := newK8sClient(options)

	return &K8s{
		options:   options,
		logger:    logger,
		clientset: clientset,
	}
}
