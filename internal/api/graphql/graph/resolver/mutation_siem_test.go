// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver_test

// Characterization tests for CreateSIEMAlert after the orchestration moved
// into the app layer. These assert the TRANSPORT responsibilities:
//   - required-fields validation fires before any app call, and
//   - buildSIEMAlertModel's field-precedence rules: persisted entity values
//     (issueVariant, componentInstance, service) win over raw input.
//
// The get-or-create logic is tested in internal/app/app.

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/resolver"
	"github.com/cloudoperators/heureka/internal/app/siem_alert"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

func fullSIEMInput() model.SIEMAlertInput {
	return model.SIEMAlertInput{
		Name:        strPtr("my-alert"),
		Service:     strPtr("my-service"),
		Region:      strPtr("eu-de-1"),
		Cluster:     strPtr("prod"),
		Namespace:   strPtr("default"),
		Pod:         strPtr("pod-1"),
		Container:   strPtr("app"),
		Description: strPtr("input-description"),
	}
}

var _ = Describe("CreateSIEMAlert (transport)", func() {
	var (
		mockApp *mocks.MockHeureka
		r       *resolver.Resolver
		ctx     context.Context
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		r = &resolver.Resolver{App: mockApp}
		ctx = context.Background()
	})

	Describe("required-fields validation", func() {
		It("rejects input missing a component-instance field before any app call", func() {
			input := fullSIEMInput()
			input.Container = nil

			result, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"CreateSIEMAlertMutationResolver: Invalid Input - service, region, cluster, namespace, pod, and container are all required",
			))
		})

		It("treats an empty-string required field as missing", func() {
			input := fullSIEMInput()
			input.Pod = strPtr("")

			result, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Invalid Input - service, region, cluster, namespace, pod, and container are all required"))
		})
	})

	Describe("response assembly (field precedence)", func() {
		It("prefers issueVariant and componentInstance values over raw input", func() {
			input := fullSIEMInput()

			orchResult := siem_alert.SIEMAlertResult{
				Service: &entity.Service{BaseService: entity.BaseService{Id: 1, CCRN: "my-service"}},
				ComponentInstance: &entity.ComponentInstance{
					Id:        10,
					Region:    "ci-region",
					Cluster:   "ci-cluster",
					Namespace: "ci-namespace",
					Pod:       "ci-pod",
					Container: "ci-container",
				},
				Issue: &entity.Issue{Id: 100, PrimaryName: "my-alert"},
				IssueVariant: &entity.IssueVariant{
					Id:          200,
					Description: "variant-description",
					Severity:    entity.Severity{Value: "High"},
					ExternalUrl: `[{"name":"ref","url":"https://example.com/ref"}]`,
				},
			}

			mockApp.On("CreateSIEMAlertFromInput", ctx, mock.AnythingOfType("siem_alert.SIEMAlertInput")).
				Return(orchResult, nil)

			alert, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(alert).NotTo(BeNil())

			Expect(alert.Name).NotTo(BeNil())
			Expect(*alert.Name).To(Equal("my-alert"))

			// issueVariant description wins over input "input-description"
			Expect(alert.Description).NotTo(BeNil())
			Expect(*alert.Description).To(Equal("variant-description"))

			Expect(alert.Severity).NotTo(BeNil())
			Expect(string(*alert.Severity)).To(Equal("High"))

			Expect(alert.Links).To(HaveLen(1))
			Expect(alert.Links[0].Name).To(Equal("ref"))
			Expect(alert.Links[0].URL).To(Equal("https://example.com/ref"))

			Expect(alert.Service).NotTo(BeNil())
			Expect(*alert.Service).To(Equal("my-service"))

			// CI values win over raw input
			Expect(*alert.Region).To(Equal("ci-region"))
			Expect(*alert.Cluster).To(Equal("ci-cluster"))
			Expect(*alert.Namespace).To(Equal("ci-namespace"))
			Expect(*alert.Pod).To(Equal("ci-pod"))
			Expect(*alert.Container).To(Equal("ci-container"))
		})

		It("falls back to input values when no entities are present", func() {
			input := fullSIEMInput()

			mockApp.On("CreateSIEMAlertFromInput", ctx, mock.AnythingOfType("siem_alert.SIEMAlertInput")).
				Return(siem_alert.SIEMAlertResult{}, nil)

			alert, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(alert.Name).To(BeNil())
			Expect(alert.Description).NotTo(BeNil())
			Expect(*alert.Description).To(Equal("input-description"))
			Expect(*alert.Service).To(Equal("my-service"))
			Expect(*alert.Region).To(Equal("eu-de-1"))
		})
	})

	Describe("error mapping", func() {
		It("wraps a SIEMError with the resolver prefix", func() {
			input := fullSIEMInput()

			mockApp.On("CreateSIEMAlertFromInput", ctx, mock.AnythingOfType("siem_alert.SIEMAlertInput")).
				Return(siem_alert.SIEMAlertResult{}, &siem_alert.SIEMError{Message: "Internal Error - when creating service"})

			_, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("CreateSIEMAlertMutationResolver: Internal Error - when creating service"))
		})

		It("wraps a generic error with the resolver prefix", func() {
			input := fullSIEMInput()

			mockApp.On("CreateSIEMAlertFromInput", ctx, mock.AnythingOfType("siem_alert.SIEMAlertInput")).
				Return(siem_alert.SIEMAlertResult{}, errors.New("db down"))

			_, err := r.Mutation().CreateSIEMAlert(ctx, input)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("CreateSIEMAlertMutationResolver: db down"))
		})
	})
})
