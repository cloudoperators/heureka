// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Some usefull structs
type Scanner struct {
	Config     Config
	KubeConfig *rest.Config
	ClientSet  *kubernetes.Clientset
}

type ServiceInfo struct {
	Name string
	Pods []PodInfo
}

type PodInfo struct {
	Labels     PodLabels
	Name       string
	Containers []ContainerInfo
}

type ContainerInfo struct {
	Name      string
	Image     string
	ImageHash string
}

type PodLabels struct {
	SupportGroup string
	Owner        string
	ServiceName  string
}

func NewScanner(kubeConfig *rest.Config, clientSet *kubernetes.Clientset, cfg Config) Scanner {
	return Scanner{
		KubeConfig: kubeConfig,
		ClientSet:  clientSet,
		Config:     cfg,
	}
}

func (s *Scanner) GetNamespaces(listOptions metav1.ListOptions) ([]v1.Namespace, error) {
	namespaces, err := s.ClientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list namespaces")
	}
	return namespaces.Items, nil
}

func (s *Scanner) GetRelevantLabels(pod v1.Pod) PodLabels {
	podLabels := PodLabels{}
	for labelName, labelValue := range pod.Labels {
		switch labelName {
		case s.Config.ServiceNameLabel:
			podLabels.ServiceName = labelValue
		case s.Config.SupportGroupLabel:
			podLabels.SupportGroup = labelValue
		default:
			continue
		}
	}
	return podLabels
}

func (s *Scanner) extractImageHash(image string) string {
	// Implementation to extract image hash
	// This is a placeholder and needs to be implemented based on your specific requirements
	return "placeholder-hash"
}

func (s *Scanner) GetPodInfo(pod v1.Pod) PodInfo {
	podInfo := PodInfo{
		Labels:     s.GetRelevantLabels(pod),
		Containers: make([]ContainerInfo, 0, len(pod.Spec.Containers)),
	}

	for _, container := range pod.Spec.Containers {
		imageHash := s.extractImageHash(container.Image)
		podInfo.Containers = append(podInfo.Containers, ContainerInfo{
			Name:      container.Name,
			Image:     container.Image,
			ImageHash: imageHash,
		})
	}

	return podInfo
}

func (s *Scanner) GetPodsAllNamespaces() []v1.Pod {
	return []v1.Pod{}
}

func (s *Scanner) GetPodsByNamespace(namespace string, listOptions metav1.ListOptions) ([]v1.Pod, error) {
	pods, err := s.ClientSet.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("couldn't list pods")
	}
	return pods.Items, nil
}
