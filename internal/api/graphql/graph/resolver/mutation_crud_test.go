// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver_test

// Characterization tests for the boilerplate CRUD mutation resolvers.
//
// These lock the CURRENT, client-observable behavior that the planned
// generic-CRUD-helper refactor (issue #1301) must preserve byte-for-byte:
//   - the exact per-resolver error string, including the leading operation
//     tag produced by baseResolver.NewResolverError ("<OpName>: <msg>"), and
//   - the baseResolver.ParseCursor failure path on a non-numeric id.
//
// They intentionally do NOT assert on the structured ToGraphQLError wording:
// migrating these messages is a separate, gated, client-observable change
// (plan Phase 3), and these assertions are expected to change ONLY then.

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mock "github.com/stretchr/testify/mock"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/resolver"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/mocks"
)

func strPtr(s string) *string { return &s }

var _ = Describe("CRUD mutation resolvers (characterization)", func() {
	var (
		mockApp *mocks.MockHeureka
		r       *resolver.Resolver
		ctx     context.Context
		input   model.UserInput
	)

	BeforeEach(func() {
		mockApp = mocks.NewMockHeureka(GinkgoT())
		r = &resolver.Resolver{App: mockApp}
		ctx = context.Background()
		input = model.UserInput{
			UniqueUserID: strPtr("I123456"),
			Name:         strPtr("Jane Doe"),
			Email:        strPtr("jane@example.com"),
			Type:         strPtr("1"),
		}
	})

	Describe("CreateUser", func() {
		It("returns the created user on success", func() {
			mockApp.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).
				Return(&entity.User{Id: 7, UniqueUserID: "I123456", Name: "Jane Doe", Email: "jane@example.com"}, nil)

			result, err := r.Mutation().CreateUser(ctx, input)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.ID).To(Equal("7"))
		})

		It("returns the exact tagged error string when the app layer fails", func() {
			mockApp.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).
				Return((*entity.User)(nil), errors.New("boom"))

			result, err := r.Mutation().CreateUser(ctx, input)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			// Current behavior: the underlying error is swallowed and a static,
			// op-tagged message is returned. The refactor must preserve this.
			Expect(err.Error()).To(Equal("CreateUserMutationResolver: Internal Error - when creating user"))
		})
	})

	Describe("UpdateUser", func() {
		It("fails on a non-numeric id via ParseCursor with the update op tag", func() {
			// ParseCursor fails before any app call, so no mock expectation is set.
			result, err := r.Mutation().UpdateUser(ctx, "not-a-number", input)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("UpdateUserMutationResolver: Internal Error - when updating user"))
		})

		It("returns the exact tagged error string when the app layer fails", func() {
			mockApp.On("UpdateUser", ctx, mock.AnythingOfType("*entity.User")).
				Return((*entity.User)(nil), errors.New("boom"))

			result, err := r.Mutation().UpdateUser(ctx, "42", input)

			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("UpdateUserMutationResolver: Internal Error - when updating user"))
		})
	})

	Describe("DeleteUser", func() {
		It("returns the id on success", func() {
			mockApp.On("DeleteUser", ctx, int64(42)).Return(nil)

			result, err := r.Mutation().DeleteUser(ctx, "42")

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("42"))
		})

		It("fails on a non-numeric id via ParseCursor with the delete op tag", func() {
			result, err := r.Mutation().DeleteUser(ctx, "not-a-number")

			Expect(result).To(Equal(""))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("DeleteUserMutationResolver: Internal Error - when deleting user"))
		})

		It("returns the exact tagged error string when the app layer fails", func() {
			mockApp.On("DeleteUser", ctx, int64(42)).Return(errors.New("boom"))

			result, err := r.Mutation().DeleteUser(ctx, "42")

			Expect(result).To(Equal(""))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("DeleteUserMutationResolver: Internal Error - when deleting user"))
		})
	})
})
