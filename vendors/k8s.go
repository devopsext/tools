package vendors

import (
	"context"
	"fmt"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
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

type K8sResourceDescribeOptions struct {
	K8sResourceOptions
}

type K8sResourceDeleteOptions struct {
	K8sResourceOptions
}

type K8sResourceScaleOptions struct {
	K8sResourceOptions
	Replicas    int
	WaitTimeout int
	PollTimeout int
}

type K8sResourceRestartOptions struct {
	K8sResourceOptions
	WaitTimeout int
	PollTimeout int
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

	var config *rest.Config

	if utils.FileExists(options.Config) {
		c, err := clientcmd.BuildConfigFromFlags("", options.Config)
		if err != nil {
			return nil, err
		}
		config = c
	} else {
		ocf, err := clientcmd.NewClientConfigFromBytes([]byte(options.Config))
		if err != nil {
			return nil, err
		}
		c, err := ocf.ClientConfig()
		if err != nil {
			return nil, err
		}
		config = c
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

func (k *K8s) resourceDescribe(clientset *kubernetes.Clientset, ctx context.Context, describeOptions K8sResourceDescribeOptions) rest.Result {

	opts := metav1.GetOptions{}

	var rClient rest.Interface

	switch describeOptions.Kind {
	case K8sResourceDeployments:
		rClient = clientset.AppsV1().RESTClient()
	default:
		rClient = clientset.CoreV1().RESTClient()
	}

	r := rClient.Get().
		UseProtobufAsDefaultIfPreferred(false).
		NamespaceIfScoped(describeOptions.Namespace, !utils.IsEmpty(describeOptions.Namespace)).
		Resource(describeOptions.Kind).
		Name(describeOptions.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx)

	return r
}

func (k *K8s) CustomResourceDescribe(options K8sOptions, describeOptions K8sResourceDescribeOptions) ([]byte, error) {

	clientset, ctx, cancel, err := k.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	return k.resourceDescribe(clientset, ctx, describeOptions).Raw()
}

func (k *K8s) ResourceDescribe(options K8sResourceDescribeOptions) ([]byte, error) {
	return k.CustomResourceDescribe(k.options, options)
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

func (k *K8s) getResourceReplicas(clientset *kubernetes.Clientset, ctx context.Context, describeOptions K8sResourceDescribeOptions) (int32, interface{}, error) {

	rGet := k.resourceDescribe(clientset, ctx, describeOptions)
	err := rGet.Error()
	if err != nil {
		return 0, nil, err
	}

	var obj interface{}
	var currentReplicas int32

	switch describeOptions.Kind {
	case K8sResourceDeployments:

		var d appsv1.Deployment
		err = rGet.Into(&d)
		if err != nil {
			return 0, nil, err
		}
		currentReplicas = *d.Spec.Replicas
		obj = &d

	case K8sResourceReplicaSets:

		var rs appsv1.ReplicaSet
		err = rGet.Into(&rs)
		if err != nil {
			return 0, nil, err
		}
		currentReplicas = *rs.Spec.Replicas
		obj = &rs

	case K8sResourceStatefulSets:

		var ss appsv1.StatefulSet
		err = rGet.Into(&ss)
		if err != nil {
			return 0, nil, err
		}
		currentReplicas = *ss.Spec.Replicas
		obj = &ss

	default:
		return 0, nil, fmt.Errorf("unsupported kind for replicas: %s", describeOptions.Kind)
	}

	return currentReplicas, obj, nil
}

func (k *K8s) replicasAreReady(obj interface{}, kind string, replicas int32) bool {

	switch kind {
	case K8sResourceDeployments:

		d := obj.(*appsv1.Deployment)

		return d.Status.ObservedGeneration >= d.Generation &&
			d.Status.ReadyReplicas == replicas &&
			d.Status.AvailableReplicas == replicas

	case K8sResourceReplicaSets:

		rs := obj.(*appsv1.ReplicaSet)
		return rs.Status.ObservedGeneration >= rs.Generation &&
			rs.Status.ReadyReplicas == replicas &&
			rs.Status.AvailableReplicas == replicas

	case K8sResourceStatefulSets:

		ss := obj.(*appsv1.StatefulSet)
		return ss.Status.ObservedGeneration >= ss.Generation &&
			ss.Status.ReadyReplicas == replicas &&
			ss.Status.AvailableReplicas == replicas

	}
	return false
}

func (k *K8s) resourceScale(options K8sOptions, scaleOptions K8sResourceScaleOptions) (*rest.Result, error) {

	clientset, ctx, cancel, err := k.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	describeOpts := K8sResourceDescribeOptions{
		K8sResourceOptions: scaleOptions.K8sResourceOptions,
	}

	desiredReplicasInt32 := int32(scaleOptions.Replicas)
	currentReplicas, obj, err := k.getResourceReplicas(clientset, ctx, describeOpts)
	if err != nil {
		return nil, err
	}
	if currentReplicas == desiredReplicasInt32 {
		return nil, fmt.Errorf("current replicas are equal to desired replicas")
	}

	switch scaleOptions.Kind {
	case K8sResourceDeployments:
		d := obj.(*appsv1.Deployment)
		d.Spec.Replicas = &desiredReplicasInt32

	case K8sResourceReplicaSets:
		rs := obj.(*appsv1.ReplicaSet)
		rs.Spec.Replicas = &desiredReplicasInt32

	case K8sResourceStatefulSets:

		ss := obj.(*appsv1.StatefulSet)
		ss.Spec.Replicas = &desiredReplicasInt32

	default:
		return nil, fmt.Errorf("unsupported kind for scaling: %s", scaleOptions.Kind)
	}

	updateOpts := metav1.UpdateOptions{}
	rClient := clientset.AppsV1().RESTClient()

	rPut := rClient.Put().
		UseProtobufAsDefaultIfPreferred(false).
		NamespaceIfScoped(scaleOptions.Namespace, !utils.IsEmpty(scaleOptions.Namespace)).
		Resource(scaleOptions.Kind).
		Name(scaleOptions.Name).
		VersionedParams(&updateOpts, scheme.ParameterCodec).
		Body(obj).
		Do(ctx)

	err = rPut.Error()
	if err != nil {
		return nil, err
	}

	if scaleOptions.WaitTimeout <= 0 {
		return &rPut, nil
	}

	// wait until scaled
	waitTimeout := time.Duration(scaleOptions.WaitTimeout) * time.Second
	pollTimeout := time.Duration(scaleOptions.PollTimeout) * time.Second

	err = wait.PollUntilContextTimeout(ctx, pollTimeout, waitTimeout, true, func(context.Context) (done bool, err error) {

		reps, obj, err := k.getResourceReplicas(clientset, ctx, describeOpts)
		if err != nil {
			return false, err
		}
		if k.replicasAreReady(obj, scaleOptions.K8sResourceOptions.Kind, desiredReplicasInt32) && reps == desiredReplicasInt32 {
			return true, nil // Polling is done, scaled to 0
		}
		return false, nil // Continue polling
	})
	if err != nil {
		return nil, fmt.Errorf("timeout waiting for resource to scale to %d: %v", desiredReplicasInt32, err)
	}

	if err != nil {
		return &rPut, err
	}

	return &rPut, nil
}

func (k *K8s) CustomResourceScale(options K8sOptions, scaleOptions K8sResourceScaleOptions) ([]byte, error) {

	rPut, err := k.resourceScale(options, scaleOptions)
	if err != nil {
		return nil, err
	}
	return rPut.Raw()
}

func (k *K8s) ResourceScale(options K8sResourceScaleOptions) ([]byte, error) {
	return k.CustomResourceScale(k.options, options)
}

func (k *K8s) CustomResourceRestart(options K8sOptions, restartOptions K8sResourceRestartOptions) ([]byte, error) {

	clientset, ctx, cancel, err := k.getClientCtx(options)
	if err != nil {
		return nil, err
	}
	defer cancel()

	// get current replicas
	describeOpts := K8sResourceDescribeOptions{
		K8sResourceOptions: restartOptions.K8sResourceOptions,
	}

	oldReplicas, _, err := k.getResourceReplicas(clientset, ctx, describeOpts)
	if err != nil {
		return nil, err
	}
	if oldReplicas == 0 {
		return nil, fmt.Errorf("current replicas are equal to 0")
	}

	// Scale to 0
	scaleOptions := K8sResourceScaleOptions{
		K8sResourceOptions: restartOptions.K8sResourceOptions,
		Replicas:           0,
		WaitTimeout:        restartOptions.WaitTimeout,
		PollTimeout:        restartOptions.PollTimeout,
	}

	_, err = k.resourceScale(options, scaleOptions)
	if err != nil {
		return nil, err
	}

	// scale back to old replicas
	scaleOptions.Replicas = int(oldReplicas)
	rPut, err := k.resourceScale(options, scaleOptions)
	if err != nil {
		return nil, err
	}
	return rPut.Raw()
}

func (k *K8s) ResourceRestart(options K8sResourceRestartOptions) ([]byte, error) {
	return k.CustomResourceRestart(k.options, options)
}

func NewK8s(options K8sOptions, logger common.Logger) *K8s {

	clientset, _ := newK8sClient(options)

	return &K8s{
		options:   options,
		logger:    logger,
		clientset: clientset,
	}
}
