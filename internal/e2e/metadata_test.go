// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/e2e/common"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testUniqueUserId    = "1"
	testUserType        = entity.HumanUserType
	testUpdatedUserType = entity.TechnicalUserType
	testUserName        = "Joe"
	testUpdatedUserName = "Donald"
	testCreatedBy       = "Creator"
	testUpdatedBy       = "Updater"
	dbDateLayout        = "2006-01-02 15:04:05 -0700 MST"
)

func createTestUser(port string) {
	user := e2e_common.QueryCreateUser(port, e2e_common.User{Id: testUniqueUserId, Type: testUserType, Name: testUserName})
	Expect(*user.Name).To(Equal(testUserName))
	Expect(*user.UniqueUserID).To(Equal(testUniqueUserId))
	Expect(entity.UserType(user.Type)).To(Equal(testUserType))
}

func updateTestUser(port string) {
	user := e2e_common.QueryUpdateUser(port, e2e_common.User{Id: testUniqueUserId, Type: testUpdatedUserType, Name: testUpdatedUserName})
	Expect(*user.Name).To(Equal(testUpdatedUserName))
	Expect(*user.UniqueUserID).To(Equal(testUniqueUserId))
	Expect(entity.UserType(user.Type)).To(Equal(testUpdatedUserType))
}

func getTestUser(port string) model.User {
	users := e2e_common.QueryGetUser(port, testUniqueUserId)
	Expect(users.TotalCount).To(Equal(1))
	return *users.Edges[0].Node
}

var _ = Describe("Creating and updating entity via API", Label("e2e", "Entities"), func() {
	var s *server.Server
	var cfg util.Config

	BeforeEach(func() {
		_ = dbm.NewTestSchema()

		cfg = dbm.DbConfig()
		cfg.Port = util2.GetRandomFreePort()
		s = server.NewServer(cfg)

		s.NonBlockingStart()
	})

	AfterEach(func() {
		s.BlockingStop()
	})

	When("New user is created via API", func() {
		var user model.User
		BeforeEach(func() {
			createTestUser(cfg.Port)
			user = getTestUser(cfg.Port)
		})
		It("shall assign CreatedBy and CreatedAt metadata fields and shall keep nil in UpdatedBy, UpdatedAt and DeltedAt metadata fields", func() {
			Expect(entity.UserType(user.Type)).To(Equal(testUserType))
			Expect(user.Metadata).To(Not(BeNil()))
			Expect(*user.Metadata.CreatedBy).To(Equal(testCreatedBy))

			createdAt, err := time.Parse(dbDateLayout, *user.Metadata.CreatedAt)
			Expect(err).Should(BeNil())
			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))

			Expect(*user.Metadata.UpdatedBy).To(BeEmpty())

			updatedAt, err := time.Parse(dbDateLayout, *user.Metadata.UpdatedAt)
			Expect(err).Should(BeNil())
			Expect(updatedAt).To(Equal(createdAt))
		})
	})
	When("User is updated via API", func() {
		var user model.User
		BeforeEach(func() {
			createTestUser(cfg.Port)
			time.Sleep(1100 * time.Millisecond)
			updateTestUser(cfg.Port)
			user = getTestUser(cfg.Port)
		})
		It("shall assign UpdatedBy and UpdatedAt metadata fields and shall keep nil in DeletedAt metadata field", func() {
			Expect(entity.UserType(user.Type)).To(Equal(testUpdatedUserType))
			Expect(user.Metadata).To(Not(BeNil()))
			Expect(*user.Metadata.CreatedBy).To(Equal(testCreatedBy))

			createdAt, err := time.Parse(dbDateLayout, *user.Metadata.CreatedAt)
			Expect(err).Should(BeNil())
			Expect(createdAt).Should(BeTemporally("~", time.Now().UTC(), 3*time.Second))

			Expect(*user.Metadata.UpdatedBy).To(Equal(testUpdatedBy))

			updatedAt, err := time.Parse(dbDateLayout, *user.Metadata.UpdatedAt)
			Expect(err).Should(BeNil())
			Expect(updatedAt).Should(BeTemporally("~", time.Now().UTC(), 2*time.Second))
			Expect(updatedAt).Should(BeTemporally(">", createdAt))
		})
	})

})
