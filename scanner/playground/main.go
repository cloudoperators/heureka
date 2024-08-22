// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"github.com/jedib0t/go-pretty/v6/table"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc" // required for OIDC support
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path"
)

func localKubeConfig() *rest.Config {
	// path to kubeconfig
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", path.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// load config from kube config
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func oidcBasedConfig() *rest.Config {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		//replace path with a kubeconfig that has a valid oidc token for your cluster
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path.Join(homedir.HomeDir(), "Library", "Application Support", "SAPCC", "u8s", ".kube", "config")},
		//replace with the context you want to use
		&clientcmd.ConfigOverrides{CurrentContext: "qa-de-1"},
	).ClientConfig()

	if err != nil {
		panic(err.Error())
	}
	return config
}

func localSAConfig() *rest.Config {

	tokenFilePath := "my_token" // path to the token file
	token, err := os.ReadFile(tokenFilePath)
	if err != nil {
		panic(err.Error())
	}

	config := &rest.Config{
		Host: "https://localhost:6443", // The API server's URL
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // allow insecure connections ( ok as we are local )
			//CAFile:   "/path/to/ca.crt", // Path to the CA certificate
		},
		BearerToken: string(token), //the token
	}
	return config
}

func main() {
	// get config
	config := localSAConfig()

	// create k8s client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// List pods from all namespaces
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Namespace", "Pod Name", "Container Name", "Image", "Container Ready", "Pod Status"})
	t.SetStyle(table.StyleColoredDark)
	for _, pod := range pods.Items {
		for _, containerS := range pod.Status.ContainerStatuses {
			t.AppendRow([]interface{}{pod.Namespace, pod.Name, containerS.Name, containerS.Image, containerS.Ready, pod.Status.Phase})
		}
	}

	t.Render()
}
