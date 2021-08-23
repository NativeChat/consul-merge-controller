/*
Copyright 2021 Progress Software Corporation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/NativeChat/consul-merge-controller/controllers/service"
	"k8s.io/client-go/rest"
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

// StartControllers ...
func StartControllers(k8sClient client.Client) {
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

	consulServiceIntentionsSource := &service.ConsulServiceIntentionsSourceReconciler{
		Client: k8sClient,
		Log:    ctrl.Log.WithName("controllers").WithName("service").WithName("ConsulServiceIntentionsSource"),
		Scheme: mgr.GetScheme(),
	}

	err = consulServiceIntentionsSource.SetupWithManager(mgr)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	go func() {
		defer ginkgo.GinkgoRecover()

		err = mgr.Start(ctrl.SetupSignalHandler())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}()
}

// CreateAndSetTestKubeconfig ...
func CreateAndSetTestKubeconfig(cfg *rest.Config) error {
	kubeconfig := apiv1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: "default",
		Clusters:       []apiv1.NamedCluster{{Name: "default", Cluster: apiv1.Cluster{Server: cfg.Host, CertificateAuthorityData: cfg.CAData}}},
		Contexts:       []apiv1.NamedContext{{Name: "default", Context: apiv1.Context{Cluster: "default", Namespace: DefaultK8sNamespace, AuthInfo: "admin"}}},
		AuthInfos:      []apiv1.NamedAuthInfo{{Name: "admin", AuthInfo: apiv1.AuthInfo{Username: "admin", ClientCertificateData: cfg.CertData, ClientKeyData: cfg.KeyData}}},
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

// StartConsulLocalEnv ...
func StartConsulLocalEnv(config *rest.Config) error {
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
		makeHackCmd := exec.Command("make", "setup-local-consul-test-env", "-f", os.Getenv("MAKEFILE_PATH"))

		if showConsulLogs {
			makeHackCmd.Stdout = os.Stdout
			makeHackCmd.Stderr = os.Stderr
		}

		err := makeHackCmd.Run()
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

// StopConsulLocalEnv ...
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
