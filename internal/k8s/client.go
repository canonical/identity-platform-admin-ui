// Copyright 2024 Canonical Ltd.
// SPDX-License-Identifier: AGPL-3.0

package k8s

import (
	"k8s.io/client-go/kubernetes"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewCoreV1Client(kubeconfig string) (coreV1.CoreV1Interface, error) {
	// httpClient := new(http.Client)
	// httpClient.Transport = otelhttp.NewTransport(http.DefaultTransport)

	var config *rest.Config
	var err error

	// use the current context in kubeconfig
	if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
		return nil, err
	}

	// creates the clientset
	// clientset, err := kubernetes.NewForConfigAndClient(config, httpClient)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset.CoreV1(), nil
}
