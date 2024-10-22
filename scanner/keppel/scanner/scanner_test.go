package scanner_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"

	"github.com/cloudoperators/heureka/scanners/keppel/scanner"
)

func TestScanner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scanner.Keppel Scanner Suite")
}

var _ = Describe("Scanner", func() {
	var (
		testScanner *scanner.Scanner
	)

	BeforeEach(func() {
		testScanner = scanner.NewScanner(scanner.Config{})
	})

	Describe("ExtractImageInfo", func() {
		Context("with valid image names", func() {
			DescribeTable("should correctly parse",
				func(image string, expected scanner.ImageInfo) {
					info, err := testScanner.ExtractImageInfo(image)
					Expect(err).NotTo(HaveOccurred())
					Expect(info).To(Equal(expected))
				},
				Entry("three components without tag",
					"registry.example.com/account/repo",
					scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "repo",
					},
				),
				Entry("with organization",
					"registry.example.com/account/org/repo",
					scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "org/repo",
					},
				),
				Entry("with tag",
					"registry.example.com/account/repo:latest",
					scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "repo",
						Tag:        "latest",
					},
				),
				Entry("with organization and tag",
					"registry.example.com/account/org/repo:v1.2.3",
					scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "org/repo",
						Tag:        "v1.2.3",
					},
				),
				Entry("with complex tag",
					"registry.example.com/account/repo:tag-with-hyphen.123",
					scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "repo",
						Tag:        "tag-with-hyphen.123",
					},
				),
				Entry("with dots in registry and repo names",
					"k8s.gcr.io/account/nginx.web:1.14.2",
					scanner.ImageInfo{
						Registry:   "k8s.gcr.io",
						Account:    "account",
						Repository: "nginx.web",
						Tag:        "1.14.2",
					},
				),
			)
		})

		Context("with invalid image names", func() {
			When("image string is empty", func() {
				It("should return an error", func() {
					_, err := testScanner.ExtractImageInfo("")
					Expect(err).To(HaveOccurred())
				})
			})

			When("image has only registry", func() {
				It("should return an error", func() {
					_, err := testScanner.ExtractImageInfo("registry.example.com")
					Expect(err).To(HaveOccurred())
				})
			})

			When("image has only registry and account", func() {
				It("should return an error", func() {
					_, err := testScanner.ExtractImageInfo("registry.example.com/account")
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("with edge cases", func() {
			When("image has multiple colons", func() {
				It("should use the first colon for tag separation", func() {
					info, err := testScanner.ExtractImageInfo("registry.example.com/account/org/repo:tag:extra")
					Expect(err).NotTo(HaveOccurred())
					Expect(info).To(Equal(scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "org/repo",
						Tag:        "tag",
					}))
				})
			})

			When("image has trailing slash", func() {
				It("should handle it correctly", func() {
					info, err := testScanner.ExtractImageInfo("registry.example.com/account/repo/")
					Expect(err).NotTo(HaveOccurred())
					Expect(info).To(Equal(scanner.ImageInfo{
						Registry:   "registry.example.com",
						Account:    "account",
						Repository: "repo/",
					}))
				})
			})
		})
	})
})
