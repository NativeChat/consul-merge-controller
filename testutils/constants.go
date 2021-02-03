package testutils

import "fmt"

const (
	// ServiceGroup is the group name for the service api.
	ServiceGroup = "service.consul.k8s.nativechat.com"

	// ServiceGroupVersion is the group version for the service api.
	ServiceGroupVersion = "v1alpha1"

	// DefaultK8sNamespace is the name of the default k8s namespace which will be used in the tests.
	DefaultK8sNamespace = "default"
)

var (
	// ServiceRouterLabel is the name of the label which stores the service router name.
	ServiceRouterLabel = fmt.Sprintf("%s/service-router", ServiceGroup)

	// ServiceFinalizer is the name of the service finalizer.
	ServiceFinalizer = fmt.Sprintf("finalizer.%s", ServiceGroup)
)
