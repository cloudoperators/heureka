// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/resolver"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

func TestResolver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resolver Suite")
}

func serviceResult(id int64) *entity.List[entity.ServiceResult] {
	return &entity.List[entity.ServiceResult]{
		Elements: []entity.ServiceResult{
			{Service: &entity.Service{BaseService: entity.BaseService{Id: id}}},
		},
	}
}

func issueResult(id int64) *entity.IssueList {
	return &entity.IssueList{
		List: &entity.List[entity.IssueResult]{
			Elements: []entity.IssueResult{
				{Issue: &entity.Issue{Id: id}},
			},
		},
	}
}

func emptyComponentList() *entity.List[entity.ComponentResult] {
	return &entity.List[entity.ComponentResult]{Elements: []entity.ComponentResult{}}
}

func singleComponentList(componentType string) *entity.List[entity.ComponentResult] {
	return &entity.List[entity.ComponentResult]{
		Elements: []entity.ComponentResult{
			{Component: &entity.Component{Id: 42, Type: componentType}},
		},
	}
}

func multipleComponentList(componentType string) *entity.List[entity.ComponentResult] {
	return &entity.List[entity.ComponentResult]{
		Elements: []entity.ComponentResult{
			{Component: &entity.Component{Id: 42, Type: componentType}},
			{Component: &entity.Component{Id: 43, Type: componentType}},
		},
	}
}

var _ = Describe("CreateRemediation", func() {
	var (
		mockApp  *mocks.MockHeureka
		r        *resolver.Resolver
		ctx      context.Context
		service  = "test-service"
		vuln     = "CVE-2024-1234"
		imageVal = "registry.example.com/myimage"
		input    model.RemediationInput
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		r = &resolver.Resolver{App: mockApp}
		ctx = context.Background()
		input = model.RemediationInput{
			Service:       &service,
			Vulnerability: &vuln,
			Image:         &imageVal,
		}
	})

	Context("when the component is not found", func() {
		It("returns an error mentioning 'component'", func() {
			mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
				Return(serviceResult(1), nil)
			mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
				Return(issueResult(2), nil)
			mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
				Return(emptyComponentList(), nil)

			mutation := r.Mutation()
			_, err := mutation.CreateRemediation(ctx, input)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("component not found"))
		})
	})

	Context("when multiple components are found (ambiguous)", func() {
		Context("and the component type is containerImage", func() {
			It("returns an error mentioning 'container image'", func() {
				mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
					Return(serviceResult(1), nil)
				mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
					Return(issueResult(2), nil)
				mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
					Return(multipleComponentList("containerImage"), nil)

				mutation := r.Mutation()
				_, err := mutation.CreateRemediation(ctx, input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("container image not found"))
			})
		})

		Context("and the component type is repository", func() {
			It("returns an error mentioning 'repository'", func() {
				mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
					Return(serviceResult(1), nil)
				mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
					Return(issueResult(2), nil)
				mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
					Return(multipleComponentList("repository"), nil)

				mutation := r.Mutation()
				_, err := mutation.CreateRemediation(ctx, input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("repository not found"))
			})
		})

		Context("and the component type is virtualMachineImage", func() {
			It("returns an error mentioning 'virtual machine image'", func() {
				mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
					Return(serviceResult(1), nil)
				mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
					Return(issueResult(2), nil)
				mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
					Return(multipleComponentList("virtualMachineImage"), nil)

				mutation := r.Mutation()
				_, err := mutation.CreateRemediation(ctx, input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("virtual machine image not found"))
			})
		})
	})

	Context("when the component is found exactly once", func() {
		It("creates the remediation successfully", func() {
			mockApp.On("ListServices", ctx, mock.Anything, mock.Anything).
				Return(serviceResult(1), nil)
			mockApp.On("ListIssues", ctx, mock.Anything, mock.Anything).
				Return(issueResult(2), nil)
			mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
				Return(singleComponentList("containerImage"), nil)
			mockApp.On("CreateRemediation", ctx, mock.MatchedBy(func(_ *entity.Remediation) bool { return true })).
				Return(&entity.Remediation{Id: 99, ComponentId: 42}, nil)

			mutation := r.Mutation()
			result, err := mutation.CreateRemediation(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})
})

var _ = Describe("UpdateRemediation", func() {
	var (
		mockApp  *mocks.MockHeureka
		r        *resolver.Resolver
		ctx      context.Context
		imageVal = "registry.example.com/myimage"
		id       = "1"
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		r = &resolver.Resolver{App: mockApp}
		ctx = context.Background()
	})

	Context("when the component is not found", func() {
		It("returns an error mentioning 'component'", func() {
			input := model.RemediationInput{Image: &imageVal}

			mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
				Return(emptyComponentList(), nil)

			mutation := r.Mutation()
			_, err := mutation.UpdateRemediation(ctx, id, input)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("component not found"))
		})
	})

	Context("when multiple components are found (ambiguous)", func() {
		Context("and the component type is containerImage", func() {
			It("returns an error mentioning 'container image'", func() {
				input := model.RemediationInput{Image: &imageVal}

				mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
					Return(multipleComponentList("containerImage"), nil)

				mutation := r.Mutation()
				_, err := mutation.UpdateRemediation(ctx, id, input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("container image not found"))
			})
		})

		Context("and the component type is repository", func() {
			It("returns an error mentioning 'repository'", func() {
				input := model.RemediationInput{Image: &imageVal}

				mockApp.On("ListComponents", ctx, mock.Anything, mock.Anything).
					Return(multipleComponentList("repository"), nil)

				mutation := r.Mutation()
				_, err := mutation.UpdateRemediation(ctx, id, input)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("repository not found"))
			})
		})
	})
})
