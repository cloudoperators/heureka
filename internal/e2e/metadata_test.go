// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/server"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

const (
	testUniqueUserId    = "1"
	testUserType        = entity.HumanUserType
	testUserName        = "Joe"
	testUpdatedUserName = "Donald"
	testCreatedBy       = "Creator"
	testUpdatedBy       = "Updater"
	dbDateLayout        = "2006-01-02 15:04:05 -0700 MST"
)

var ()

func createUser(port string) {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/create.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("input", map[string]string{
		"uniqueUserId": testUniqueUserId,
		"type":         entity.GetUserTypeString(testUserType),
		"name":         testUserName,
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		User model.User `json:"createUser"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}

	Expect(*respData.User.Name).To(Equal(testUserName))
	Expect(*respData.User.UniqueUserID).To(Equal(testUniqueUserId))
	Expect(entity.UserType(respData.User.Type)).To(Equal(testUserType))
}

func updateUser(port string) {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/update.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("id", testUniqueUserId)
	req.Var("input", map[string]string{
		"name": testUpdatedUserName,
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		User model.User `json:"updateUser"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}

	Expect(*respData.User.Name).To(Equal(testUpdatedUserName))
	Expect(*respData.User.UniqueUserID).To(Equal(testUniqueUserId))
	Expect(entity.UserType(respData.User.Type)).To(Equal(testUserType))
}

func getUser(port string) model.User {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/listUsers.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("filter", map[string]string{"uniqueUserId": testUniqueUserId})
	req.Var("first", 1)
	req.Var("after", "0")

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		Users model.UserConnection `json:"Users"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	Expect(respData.Users.TotalCount).To(Equal(1))
	return *respData.Users.Edges[0].Node
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
			createUser(cfg.Port)
			user = getUser(cfg.Port)
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
			createUser(cfg.Port)
			time.Sleep(1100 * time.Millisecond)
			updateUser(cfg.Port)
			user = getUser(cfg.Port)
		})
		It("shall assign UpdatedBy and UpdatedAt metadata fields and shall keep nil in DeletedAt metadata field", func() {
			Expect(entity.UserType(user.Type)).To(Equal(testUserType))
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
