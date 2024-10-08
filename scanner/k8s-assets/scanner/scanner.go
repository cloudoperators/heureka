// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner

import (
	"context"
	"fmt"

	"strings"

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
	Name         string
	SupportGroup string
	Pods         []PodInfo
}

type PodReplicaSet struct {
	GenerateName string
	Pods         []PodInfo
}

type PodInfo struct {
	Labels       PodLabels
	Name         string
	GenerateName string
	Namespace    string
	UID          string
	Containers   []ContainerInfo
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

// GetNamespaces fetches all available namespaces for a cluster
func (s *Scanner) GetNamespaces(listOptions metav1.ListOptions) ([]v1.Namespace, error) {
	namespaces, err := s.ClientSet.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("couldn't list namespaces: %w", err)
	}
	return namespaces.Items, nil
}

// GetRelevantLabels returns only specific/relevant pod labels
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

// ParseImageHash extracts the image ID hash (after the @)
func (s *Scanner) ParseImageHash(image string) (string, error) {
	parts := strings.Split(image, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("image does not contain a manifest hash: %s", image)
	}
	return parts[1], nil
}

// fetchImageId fetches the right imageId for a specific container
func (s *Scanner) fetchImageId(pod v1.Pod, container v1.Container) string {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		statusI := strings.SplitN(containerStatus.Image, "/", 2)
		containerI := strings.SplitN(container.Image, "/", 2)
		if len(statusI) > 1 && len(containerI) > 1 {
			if statusI[1] == containerI[1] {
				return containerStatus.ImageID
			}
		} else if containerStatus.Image == container.Image {
			return containerStatus.ImageID
		}
	}
	return ""
}

func (s *Scanner) GetPodInfo(pod v1.Pod) PodInfo {
	podInfo := PodInfo{
		Name:         pod.Name,
		GenerateName: pod.GenerateName,
		UID:          string(pod.UID),
		Namespace:    pod.Namespace,
		Labels:       s.GetRelevantLabels(pod),
		Containers:   make([]ContainerInfo, 0, len(pod.Spec.Containers)),
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		podInfo.Containers = append(podInfo.Containers, ContainerInfo{
			Name:      containerStatus.Name,
			Image:     containerStatus.Image,
			ImageHash: containerStatus.ImageID,
		})
	}

	return podInfo
}

// GroupPodsByGenerateName will group pod replicas by "GenerateName"
// wll return a list pf PodReplicaSet
func (s *Scanner) GroupPodsByGenerateName(pods []v1.Pod) []PodReplicaSet {
	podGroups := make(map[string][]PodInfo)

	for _, pod := range pods {
		podInfo := s.GetPodInfo(pod)
		key := podInfo.GenerateName
		if key == "" {
			key = podInfo.Name // Use Name if GenerateName is empty
		}
		podGroups[key] = append(podGroups[key], podInfo)
	}

	result := make([]PodReplicaSet, 0, len(podGroups))
	for generateName, pods := range podGroups {
		result = append(result, PodReplicaSet{
			GenerateName: generateName,
			Pods:         pods,
		})
	}

	return result
}

// GetServiceInfo extracts meta information from a PodInfo object
func (s *Scanner) GetServiceInfo(podInfo PodInfo) ServiceInfo {
	return ServiceInfo{
		Name:         podInfo.Labels.ServiceName,
		SupportGroup: podInfo.Labels.SupportGroup,
	}
}

// GetPodsByNamespace returns a list of pods for a given namespace
func (s *Scanner) GetPodsByNamespace(namespace string, listOptions metav1.ListOptions) ([]v1.Pod, error) {
	pods, err := s.ClientSet.CoreV1().Pods(namespace).List(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("couldn't list pods: %w", err)
	}
	return pods.Items, nil
}
