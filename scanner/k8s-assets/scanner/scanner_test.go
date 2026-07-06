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
