package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/gomega"

	"github.com/NativeChat/consul-merge-controller/apis/service/v1alpha1"
	"github.com/NativeChat/consul-merge-controller/controllers/service"
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var consulCmd *exec.Cmd
var consulK8sCmd *exec.Cmd

var showConsulLogs = os.Getenv("SHOW_CONSUL_LOGS") == "true"
var assetsDir = os.Getenv("ENVTEST_ASSETS_DIR")
var testBinDir = path.Join(assetsDir, "bin")
var kubeconfigPath = path.Join(assetsDir, "kubeconfig.json")

func StartConsulServiceRouteController(k8sClient client.Client) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Scheme:             k8sClient.Scheme(),
		MetricsBindAddress: "0",
	})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	consulServiceRouterController := &service.ConsulServiceRouteReconciler{
		Client: k8sClient,
		Log:    ctrl.Log.WithName("controllers").WithName("service").WithName("ConsulServiceRoute"),
		Scheme: mgr.GetScheme(),
	}

	err = consulServiceRouterController.SetupWithManager(mgr)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	go func() {
		err = mgr.Start(ctrl.SetupSignalHandler())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}()
}

func CreateAndSetTestKubeconfig(hostWithPort string) error {
	kubeconfig := apiv1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: "default",
		Clusters:       []apiv1.NamedCluster{{Name: "default", Cluster: apiv1.Cluster{Server: fmt.Sprintf("http://%s", hostWithPort)}}},
		Contexts:       []apiv1.NamedContext{{Name: "default", Context: apiv1.Context{Cluster: "default", Namespace: DefaultK8sNamespace, AuthInfo: "admin"}}},
		AuthInfos:      []apiv1.NamedAuthInfo{{Name: "admin", AuthInfo: apiv1.AuthInfo{Username: "admin", Password: ""}}},
	}

	serializedConfig, err := json.Marshal(kubeconfig)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(kubeconfigPath, serializedConfig, 0644)
	if err != nil {
		return nil
	}

	err = os.Setenv("KUBECONFIG", kubeconfigPath)

	return err
}

func StartConsulLocalEnv() error {
	if consulCmd == nil {
		logf.Log.Info("starting local consul")

		consulCmd = exec.Command(path.Join(testBinDir, "consul"), "agent", "-dev")
		if showConsulLogs {
			consulCmd.Stdout = os.Stdout
			consulCmd.Stderr = os.Stderr
		}

		err := consulCmd.Start()
		if err != nil {
			return err
		}

		time.Sleep(time.Second)

		logf.Log.Info("started local consul")
	}

	if consulK8sCmd == nil {
		logf.Log.Info("creating consul CRDs in the current k8s cluster")
		err := exec.Command("make", "setup-consul-test-env", "-f", os.Getenv("MAKEFILE_PATH")).Run()
		if err != nil {
			return err
		}

		logf.Log.Info("created consul CRDs in the current k8s cluster")

		logf.Log.Info("starting local consul-k8s")

		consulK8sCmd = exec.Command(path.Join(testBinDir, "consul-k8s"), "controller", "-enable-webhooks=false", "-datacenter", "dc1")
		if showConsulLogs {
			consulK8sCmd.Stdout = os.Stdout
			consulK8sCmd.Stderr = os.Stderr
		}

		err = consulK8sCmd.Start()
		if err != nil {
			return err
		}

		time.Sleep(time.Second)

		logf.Log.Info("started local consul-k8s")
	}

	return nil
}

func StopConsulLocalEnv() error {
	var consulStopErr error
	var consulK8sStopErr error
	if consulCmd != nil {
		consulStopErr = consulCmd.Process.Kill()
	}

	if consulK8sCmd != nil {
		consulK8sStopErr = consulK8sCmd.Process.Kill()
	}

	if consulStopErr != nil {
		return consulStopErr
	}

	if consulK8sStopErr != nil {
		return consulK8sStopErr
	}

	return nil
}

func CreateServiceDefaults(ctx context.Context, k8sClient client.Client, service string, protocol ...string) error {
	serviceProtocol := "http"
	if len(protocol) != 0 {
		serviceProtocol = protocol[0]
	}

	sd := &consulk8s.ServiceDefaults{
		TypeMeta: v1.TypeMeta{
			APIVersion: consulk8s.SchemeBuilder.GroupVersion.Version,
			Kind:       "ServiceDefaults",
		},
		ObjectMeta: v1.ObjectMeta{Name: service, Namespace: DefaultK8sNamespace},
		Spec:       consulk8s.ServiceDefaultsSpec{Protocol: serviceProtocol},
	}

	err := k8sClient.Create(ctx, sd)

	time.Sleep(100 * time.Millisecond)

	return err
}

func CreateConsulServiceRoute(ctx context.Context, k8sClient client.Client, serviceRouter string, route consulk8s.ServiceRoute) error {
	name := route.Destination.Service

	serviceDefaults := new(consulk8s.ServiceDefaults)
	exists, err := getK8sObject(ctx, k8sClient, name, serviceDefaults)
	if err != nil {
		return err
	}

	if !exists {
		err = CreateServiceDefaults(ctx, k8sClient, name)
		if err != nil {
			return err
		}
	}

	csr := &v1alpha1.ConsulServiceRoute{
		TypeMeta: v1.TypeMeta{
			APIVersion: v1alpha1.GroupVersion.Version,
			Kind:       "ConsulServiceRoute",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: DefaultK8sNamespace,

			Labels: map[string]string{ServiceRouterLabel: serviceRouter},
		},
		Spec: v1alpha1.ConsulServiceRouteSpec{
			Route: route,
		},
	}

	err = k8sClient.Create(ctx, csr)

	time.Sleep(300 * time.Millisecond)

	return err
}

func GetConsulServiceRoute(ctx context.Context, k8sClient client.Client, service string) (*v1alpha1.ConsulServiceRoute, error) {
	csr := new(v1alpha1.ConsulServiceRoute)
	exists, err := getK8sObject(ctx, k8sClient, service, csr)
	if !exists {
		csr = nil
	}

	return csr, err
}

func GetServiceRouter(ctx context.Context, k8sClient client.Client, name string) (*consulk8s.ServiceRouter, error) {
	sr := new(consulk8s.ServiceRouter)
	exists, err := getK8sObject(ctx, k8sClient, name, sr)
	if !exists {
		sr = nil
	}

	return sr, err
}

func DeleteServiceDefaults(ctx context.Context, k8sClient client.Client, service string) error {
	sd := new(consulk8s.ServiceDefaults)
	err := deleteK8sObject(ctx, k8sClient, service, sd)

	return err
}

func DeleteConsulServiceRoute(ctx context.Context, k8sClient client.Client, name string) error {
	csr := new(v1alpha1.ConsulServiceRoute)
	err := deleteK8sObject(ctx, k8sClient, name, csr)

	return err
}

func ExpectConsulServiceRoute(ctx context.Context, k8sClient client.Client, name string, expected consulk8s.ServiceRoute) {
	csr, err := GetConsulServiceRoute(ctx, k8sClient, name)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	gomega.Expect(csr).NotTo(gomega.BeNil())
	gomega.Expect(csr.Finalizers).To(gomega.ContainElement(ServiceFinalizer))

	gomega.Expect(csr.Status.UpdatedAt).NotTo(gomega.BeEmpty())

	gomega.Expect(csr.Spec.Route).To(gomega.Equal(expected))
}

func CreateHTTPPathPrefixRoute(service, pathPrefix string) consulk8s.ServiceRoute {
	result := consulk8s.ServiceRoute{
		Match:       &consulk8s.ServiceRouteMatch{HTTP: &consulk8s.ServiceRouteHTTPMatch{PathPrefix: pathPrefix}},
		Destination: &consulk8s.ServiceRouteDestination{Service: service},
	}

	return result
}

func deleteK8sObject(ctx context.Context, k8sClient client.Client, name string, obj client.Object) error {
	exists, err := getK8sObject(ctx, k8sClient, name, obj)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	err = k8sClient.Delete(ctx, obj.(client.Object))

	return err
}

func getK8sObject(ctx context.Context, k8sClient client.Client, name string, obj client.Object) (exists bool, err error) {
	err = k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: DefaultK8sNamespace}, obj)
	if err != nil {
		if errors.IsNotFound(err) {
			obj = nil
			return false, nil
		}

		return false, err
	}

	return true, nil
}
