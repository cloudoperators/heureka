// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package processor_test

import (
	"os"

	"github.com/cloudoperators/heureka/scanners/k8s-assets/processor"
	"github.com/cloudoperators/heureka/scanners/k8s-assets/scanner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Processor", func() {
	var p *processor.Processor

	BeforeEach(func() {
		p = &processor.Processor{}
	})

	Describe("CollectUniqueContainers", func() {
		Context("with a single pod and single container", func() {
			It("should return one unique container", func() {
				podReplicaSet := scanner.PodSet{
					GenerateName: "test-pod",
					Pods: []scanner.PodInfo{
						{
							Name: "test-pod-1",
							Containers: []scanner.ContainerInfo{
								{Name: "container1", Image: "image1", ImageHash: "hash1"},
							},
						},
					},
				}

				uniqueContainers := p.CollectUniqueContainers(podReplicaSet)
				Expect(uniqueContainers).To(HaveLen(1))
				Expect(uniqueContainers[0].Name).To(Equal("container1"))
				Expect(uniqueContainers[0].ImageHash).To(Equal("hash1"))
				Expect(uniqueContainers[0].Count).To(Equal(1))
			})
		})

		Context("with multiple pods having overlapping containers", func() {
			It("should return all unique containers", func() {
				podReplicaSet := scanner.PodSet{
					GenerateName: "test-pod",
					Pods: []scanner.PodInfo{
						{
							Name: "test-pod-1",
							Containers: []scanner.ContainerInfo{
								{Name: "container1", Image: "image1", ImageHash: "hash1"},
								{Name: "container2", Image: "image2", ImageHash: "hash2"},
							},
						},
						{
							Name: "test-pod-2",
							Containers: []scanner.ContainerInfo{
								{Name: "container1", Image: "image1", ImageHash: "hash1"},
								{Name: "container3", Image: "image3", ImageHash: "hash3"},
							},
						},
					},
				}

				uniqueContainers := p.CollectUniqueContainers(podReplicaSet)
				Expect(uniqueContainers).To(HaveLen(3))
				containerNames := []string{
					uniqueContainers[0].Name,
					uniqueContainers[1].Name,
					uniqueContainers[2].Name,
				}
				Expect(containerNames).To(ConsistOf("container1", "container2", "container3"))

				// Check counts
				for _, c := range uniqueContainers {
					if c.Name == "container1" {
						Expect(c.Count).To(Equal(2))
					} else {
						Expect(c.Count).To(Equal(1))
					}
				}
			})
		})

		Context("with multiple pods having different image hashes", func() {
			It("should return containers with different image hashes and correct counts", func() {
				podReplicaSet := scanner.PodSet{
					GenerateName: "test-pod",
					Pods: []scanner.PodInfo{
						{
							Name: "test-pod-1",
							Containers: []scanner.ContainerInfo{
								{Name: "container1", Image: "image1", ImageHash: "hash1"},
							},
						},
						{
							Name: "test-pod-2",
							Containers: []scanner.ContainerInfo{
								{Name: "container1", Image: "image1", ImageHash: "hash1"},
								{Name: "container1", Image: "image1", ImageHash: "hash2"},
								{Name: "container2", Image: "image1", ImageHash: "hash2"},
								{Name: "container3", Image: "image1", ImageHash: "hash2"},
								{Name: "container4", Image: "image1", ImageHash: "hash3"},
							},
						},
					},
				}
				uniqueContainers := p.CollectUniqueContainers(podReplicaSet)
				Expect(uniqueContainers).To(HaveLen(5))

				// Helper function to find a container by name and hash
				findContainer := func(name, hash string) *processor.UniqueContainerInfo {
					for _, c := range uniqueContainers {
						if c.Name == name && c.ImageHash == hash {
							return &c
						}
					}
					return nil
				}

				// Check each unique container
				c1 := findContainer("container1", "hash1")
				Expect(c1).NotTo(BeNil())
				Expect(c1.Count).To(Equal(2))

				c2 := findContainer("container1", "hash2")
				Expect(c2).NotTo(BeNil())
				Expect(c2.Count).To(Equal(1))

				c3 := findContainer("container2", "hash2")
				Expect(c3).NotTo(BeNil())
				Expect(c3.Count).To(Equal(1))

				c4 := findContainer("container3", "hash2")
				Expect(c4).NotTo(BeNil())
				Expect(c4.Count).To(Equal(1))

				c5 := findContainer("container4", "hash3")
				Expect(c5).NotTo(BeNil())
				Expect(c5.Count).To(Equal(1))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty pods, no containers, and empty replica sets", func() {
				emptyPodReplicaSet := scanner.PodSet{
					GenerateName: "empty-pods",
					Pods:         []scanner.PodInfo{{}},
				}
				result := p.CollectUniqueContainers(emptyPodReplicaSet)
				Expect(result).To(BeEmpty())

				noContainersPodReplicaSet := scanner.PodSet{
					GenerateName: "no-containers",
					Pods: []scanner.PodInfo{
						{Name: "pod1", Containers: []scanner.ContainerInfo{}},
						{Name: "pod2", Containers: []scanner.ContainerInfo{}},
					},
				}
				result = p.CollectUniqueContainers(noContainersPodReplicaSet)
				Expect(result).To(BeEmpty())

				emptyReplicaSet := scanner.PodSet{
					GenerateName: "empty-replica-set",
					Pods:         []scanner.PodInfo{},
				}
				result = p.CollectUniqueContainers(emptyReplicaSet)
				Expect(result).To(BeEmpty())
			})

			It("should handle containers with empty names or image hashes", func() {
				podReplicaSet := scanner.PodSet{
					GenerateName: "empty-fields",
					Pods: []scanner.PodInfo{
						{Name: "pod1", Containers: []scanner.ContainerInfo{
							{Name: "", Image: "image1", ImageHash: "hash1"},
							{Name: "container2", Image: "image2", ImageHash: ""},
						}},
					},
				}
				result := p.CollectUniqueContainers(podReplicaSet)
				Expect(result).To(HaveLen(2))
				Expect(result[0].Count).To(Equal(1))
				Expect(result[1].Count).To(Equal(1))
			})

			It("should handle when all containers across all pods are identical", func() {
				container := scanner.ContainerInfo{
					Name:      "container",
					Image:     "image",
					ImageHash: "hash",
				}
				podReplicaSet := scanner.PodSet{
					GenerateName: "all-identical",
					Pods: []scanner.PodInfo{
						{Name: "pod1", Containers: []scanner.ContainerInfo{container, container}},
						{Name: "pod2", Containers: []scanner.ContainerInfo{container, container}},
					},
				}
				result := p.CollectUniqueContainers(podReplicaSet)
				Expect(result).To(HaveLen(1))
				Expect(result[0].Count).To(Equal(4))
			})
		})
	})
})

var _ = Describe("AdvancedConfig", func() {
	Describe("IsExcludedPod", func() {
		Context("with an exclusion list", func() {
			var cfg *processor.AdvancedConfig

			BeforeEach(func() {
				cfg = &processor.AdvancedConfig{
					ExcludedPods: []string{"keep-image-pulled", "debug-pod"},
				}
			})

			It("returns true for an exact match", func() {
				Expect(cfg.IsExcludedPod("keep-image-pulled")).To(BeTrue())
				Expect(cfg.IsExcludedPod("debug-pod")).To(BeTrue())
			})

			It("returns true when generateName has a trailing dash (Deployment/DaemonSet pods)", func() {
				Expect(cfg.IsExcludedPod("keep-image-pulled-")).To(BeTrue())
				Expect(cfg.IsExcludedPod("debug-pod-")).To(BeTrue())
			})

			It("returns true for a standalone pod whose Name matches exactly (no trailing dash)", func() {
				Expect(cfg.IsExcludedPod("keep-image-pulled")).To(BeTrue())
			})

			It("returns false for a pod not in the list", func() {
				Expect(cfg.IsExcludedPod("my-app")).To(BeFalse())
			})

			It("returns false for a partial match", func() {
				Expect(cfg.IsExcludedPod("keep-image-pulled-extra")).To(BeFalse())
				Expect(cfg.IsExcludedPod("keep-image")).To(BeFalse())
			})

			It("returns false for an empty string", func() {
				Expect(cfg.IsExcludedPod("")).To(BeFalse())
			})
		})

		Context("with an empty exclusion list", func() {
			It("always returns false", func() {
				cfg := &processor.AdvancedConfig{ExcludedPods: []string{}}
				Expect(cfg.IsExcludedPod("keep-image-pulled")).To(BeFalse())
			})
		})
	})

	Describe("NewAdvancedConfig", func() {
		It("parses excluded_pods from a YAML file", func() {
			f, err := os.CreateTemp("", "advanced-config-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(f.Name())

			_, err = f.WriteString("excluded_pods:\n  - keep-image-pulled\n  - debug-pod\n")
			Expect(err).NotTo(HaveOccurred())
			f.Close()

			cfg, err := processor.NewAdvancedConfig(f.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.ExcludedPods).To(ConsistOf("keep-image-pulled", "debug-pod"))
			Expect(cfg.IsExcludedPod("keep-image-pulled")).To(BeTrue())
			Expect(cfg.IsExcludedPod("other-pod")).To(BeFalse())
		})

		It("returns an empty exclusion list when the file does not exist", func() {
			cfg, err := processor.NewAdvancedConfig("/nonexistent/path/config.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.ExcludedPods).To(BeEmpty())
		})

		It("returns an empty exclusion list when excluded_pods is absent from the file", func() {
			f, err := os.CreateTemp("", "advanced-config-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			defer os.Remove(f.Name())

			_, err = f.WriteString("side_cars:\n  - name: linkerd-proxy\n    service: linkerd\n    support_group: containers\n")
			Expect(err).NotTo(HaveOccurred())
			f.Close()

			cfg, err := processor.NewAdvancedConfig(f.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.ExcludedPods).To(BeEmpty())
		})
	})
})
