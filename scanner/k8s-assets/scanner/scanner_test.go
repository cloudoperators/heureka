// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package scanner_test

import (
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Scanner", func() {
	var s scanner.Scanner

	BeforeEach(func() {
		s = scanner.Scanner{}
	})

	Describe("GetPodInfo", func() {
		Context("pod with both regular and init containers (sidecar pattern)", func() {
			It("returns ContainerInfo entries for all containers including init sidecars", func() {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:         "unbound2-5546bc894-kwdjr",
						GenerateName: "unbound2-5546bc894-",
						OwnerReferences: []metav1.OwnerReference{
							{Kind: "ReplicaSet", Name: "unbound2-5546bc894"},
						},
					},
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name:    "unbound",
								Image:   "keppel.eu-de-1.cloud.sap/ccloud/unbound:20260625084743",
								ImageID: "keppel.eu-de-1.cloud.sap/ccloud/unbound@sha256:af51ddd94eb85b2e380eaca8e443ece1e08a3ebb5f1a5faba871e0ba8c787ce8",
							},
						},
						InitContainerStatuses: []v1.ContainerStatus{
							{
								Name:    "metric",
								Image:   "keppel.eu-de-1.cloud.sap/ccloud/unbound_exporter:20260615060514",
								ImageID: "keppel.eu-de-1.cloud.sap/ccloud/unbound_exporter@sha256:6128f0d64f4299873b6f0c3b05a5dd1c95272be21d27b0594043d1021a3c94ec",
							},
							{
								Name:    "dnstap",
								Image:   "keppel.eu-de-1.cloud.sap/ccloud/dnstap:20260615060514",
								ImageID: "keppel.eu-de-1.cloud.sap/ccloud/dnstap@sha256:40d567e3ca6d80ebaead2c4c0672f597a1aad0243e7ca38b23e375233ff8c204",
							},
							{
								Name:    "bind-rpz-proxy",
								Image:   "keppel.eu-de-1.cloud.sap/ccloud/bind-rpz-proxy:20260615060514",
								ImageID: "keppel.eu-de-1.cloud.sap/ccloud/bind-rpz-proxy@sha256:2200d68991754ae04ff1096add2af78be164f9c6328eba90ccb81e610f3e6e9d",
							},
						},
					},
				}

				podInfo := s.GetPodInfo(pod)

				Expect(podInfo.Containers).To(HaveLen(4))

				names := make([]string, 0, len(podInfo.Containers))
				for _, c := range podInfo.Containers {
					names = append(names, c.Name)
				}
				Expect(names).To(ConsistOf("unbound", "metric", "dnstap", "bind-rpz-proxy"))
			})
		})

		Context("pod with only regular containers", func() {
			It("returns ContainerInfo entries for all regular containers", func() {
				pod := v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name:    "app",
								Image:   "keppel.eu-de-1.cloud.sap/ccloud/app:latest",
								ImageID: "keppel.eu-de-1.cloud.sap/ccloud/app@sha256:aabbcc",
							},
						},
					},
				}

				podInfo := s.GetPodInfo(pod)

				Expect(podInfo.Containers).To(HaveLen(1))
				Expect(podInfo.Containers[0].Name).To(Equal("app"))
			})
		})
	})

	Describe("GroupPodsByGenerateName", func() {
		Context("standalone pod (no controller, empty GenerateName)", func() {
			It("uses Name as the group key", func() {
				pods := []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:         "keep-image-pulled",
							GenerateName: "",
						},
					},
				}

				groups := s.GroupPodsByGenerateName(pods)
				Expect(groups).To(HaveLen(1))
				Expect(groups[0].GenerateName).To(Equal("keep-image-pulled"))
			})
		})

		Context("managed pod (Deployment/DaemonSet, GenerateName set by controller)", func() {
			It("uses GenerateName as the group key and groups replicas together", func() {
				pods := []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:         "keep-image-pulled-7d9f8b-xk2pf",
							GenerateName: "keep-image-pulled-",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "ReplicaSet", Name: "keep-image-pulled-7d9f8b"},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:         "keep-image-pulled-7d9f8b-ab3cd",
							GenerateName: "keep-image-pulled-",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "ReplicaSet", Name: "keep-image-pulled-7d9f8b"},
							},
						},
					},
				}

				groups := s.GroupPodsByGenerateName(pods)
				Expect(groups).To(HaveLen(1))
				Expect(groups[0].GenerateName).To(Equal("keep-image-pulled-"))
				Expect(groups[0].Pods).To(HaveLen(2))
			})
		})

		Context("Job pod", func() {
			It("uses the Job base name (stripped of run suffix) as the group key", func() {
				pods := []v1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:         "my-job-28123456-xk2pf",
							GenerateName: "my-job-28123456-",
							OwnerReferences: []metav1.OwnerReference{
								{Kind: "Job", Name: "my-job-28123456"},
							},
						},
					},
				}

				groups := s.GroupPodsByGenerateName(pods)
				Expect(groups).To(HaveLen(1))
				Expect(groups[0].GenerateName).To(Equal("my-job"))
			})
		})
	})
})
