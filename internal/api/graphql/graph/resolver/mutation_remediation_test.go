// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver_test

// Characterization tests for the Remediation resolvers after the orchestration
// moved into the app layer. These now assert the TRANSPORT responsibilities:
//   - the resolver delegates to App.CreateRemediationFromInput /
//     UpdateRemediationFromInput, and
//   - it maps the domain RemediationReferenceError back to the exact,
//     client-facing error strings (including the component friendly-name).
// The cross-domain reference-resolution branches are tested in
// internal/app/app.

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/resolver"
	"github.com/cloudoperators/heureka/internal/app/remediation"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

var _ = Describe("CreateRemediation (transport mapping)", func() {
	var (
		mockApp *mocks.MockHeureka
		r       *resolver.Resolver
		ctx     context.Context
		input   model.RemediationInput
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		r = &resolver.Resolver{App: mockApp}
		ctx = context.Background()
		input = model.RemediationInput{
			Service:       strPtr("test-service"),
			Vulnerability: strPtr("CVE-2024-1234"),
			Image:         strPtr("registry.example.com/img"),
		}
	})

	It("returns the created remediation on success", func() {
		mockApp.On("CreateRemediationFromInput", ctx, mock.AnythingOfType("remediation.RemediationCreateInput")).
			Return(&entity.Remediation{Id: 99}, nil)

		result, err := r.Mutation().CreateRemediation(ctx, input)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
	})

	DescribeTable(
		"maps a domain reference error to the exact client message",
		func(refErr *remediation.RemediationReferenceError, expected string) {
			mockApp.On("CreateRemediationFromInput", ctx, mock.AnythingOfType("remediation.RemediationCreateInput")).
				Return((*entity.Remediation)(nil), refErr)

			_, err := r.Mutation().CreateRemediation(ctx, input)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(expected))
		},
		Entry("service",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceService},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - service id not found"),
		Entry("issue",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceIssue},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - issue id not found"),
		Entry("component not found",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceComponent},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - component not found"),
		Entry("ambiguous containerImage",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceComponent, Ambiguous: true, ComponentType: "containerImage"},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - container image not found"),
		Entry("ambiguous repository",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceComponent, Ambiguous: true, ComponentType: "repository"},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - repository not found"),
		Entry("ambiguous virtualMachineImage",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceComponent, Ambiguous: true, ComponentType: "virtualMachineImage"},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - virtual machine image not found"),
		Entry("user",
			&remediation.RemediationReferenceError{Reference: remediation.RemediationReferenceUser},
			"CreateRemediationMutationResolver: Internal Error - when creating remediation - user id not found"),
	)

	It("maps a non-reference (persistence) error to the generic create message", func() {
		mockApp.On("CreateRemediationFromInput", ctx, mock.AnythingOfType("remediation.RemediationCreateInput")).
			Return((*entity.Remediation)(nil), errors.New("db down"))

		_, err := r.Mutation().CreateRemediation(ctx, input)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("CreateRemediationMutationResolver: Internal Error - when creating remediation"))
	})
})

var _ = Describe("UpdateRemediation (transport mapping)", func() {
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

	It("fails on a non-numeric id before any app call", func() {
		_, err := r.Mutation().UpdateRemediation(ctx, "not-a-number", model.RemediationInput{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("UpdateRemediationMutationResolver: Internal Error - when updating remediation"))
	})

	It("maps an ambiguous component error to the friendly client message", func() {
		mockApp.On("UpdateRemediationFromInput", ctx, mock.AnythingOfType("remediation.RemediationUpdateInput")).
			Return((*entity.Remediation)(nil), &remediation.RemediationReferenceError{
				Reference: remediation.RemediationReferenceComponent, Ambiguous: true, ComponentType: "containerImage",
			})

		_, err := r.Mutation().UpdateRemediation(ctx, "1", model.RemediationInput{Image: strPtr("img")})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("UpdateRemediationMutationResolver: Internal Error - when updating remediation - container image not found"))
	})

	It("returns the updated remediation on success", func() {
		mockApp.On("UpdateRemediationFromInput", ctx, mock.AnythingOfType("remediation.RemediationUpdateInput")).
			Return(&entity.Remediation{Id: 1}, nil)

		result, err := r.Mutation().UpdateRemediation(ctx, "1", model.RemediationInput{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
	})
})
