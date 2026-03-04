// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	"github.com/cloudoperators/heureka/internal/util"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

var (
	systemUserName         = "systemuser"
	systemUserUniqueUserId = "S0000000"
)

type Number interface {
	int | int64
}

func SubtractSystemUsers[T Number](n T) T {
	return n - 1
}

func SubtractSystemUserName(v []*string) []*string {
	return lo.Filter(v, func(val *string, _ int) bool {
		return val == nil || *val != systemUserName
	})
}

func SubtractSystemUserNameFromValueItems(v []*model.ValueItem) []*model.ValueItem {
	return lo.Filter(v, func(val *model.ValueItem, _ int) bool {
		return val == nil || *val.Value != systemUserUniqueUserId
	})
}

func SubtractSystemUserUniqueUserId(v []*string) []*string {
	return lo.Filter(v, func(val *string, _ int) bool {
		return val == nil || *val != systemUserUniqueUserId
	})
}

func SubtractSystemUserUniqueUserIdVL(v []string) []string {
	return lo.Filter(v, func(val string, _ int) bool {
		return val != systemUserUniqueUserId
	})
}

func SubtractSystemUserNameVL(v []string) []string {
	return lo.Filter(v, func(val string, _ int) bool {
		return val != systemUserName
	})
}

func SubtractSystemUsersEntity(v []entity.UserResult) []entity.UserResult {
	return lo.Filter(v, func(val entity.UserResult, _ int) bool {
		return val.User.UniqueUserID != systemUserUniqueUserId
	})
}

func SubtractSystemUserId(v []int64) []int64 {
	return lo.Filter(v, func(val int64, _ int) bool {
		return val != util.SystemUserId
	})
}

func ExpectNonSystemUserCount(n, expectedN int) {
	Expect(SubtractSystemUsers(n)).To(Equal(expectedN))
}

func ExpectNonSystemUserNames(v, expectedV []*string) {
	Expect(SubtractSystemUserName(v)).To(Equal(expectedV))
}

func ExpectNonSystemUserUniqueUserIds(v, expectedV []*string) {
	Expect(SubtractSystemUserUniqueUserId(v)).To(Equal(expectedV))
}

type User struct {
	UniqueUserID string
	Type         entity.UserType
	Name         string
	Email        string
}

func QueryCreateUser(port string, user User) *model.User {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/create.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("input", map[string]string{
		"uniqueUserId": user.UniqueUserID,
		"type":         entity.GetUserTypeString(user.Type),
		"name":         user.Name,
		"email":        user.Email,
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		User model.User `json:"createUser"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return &respData.User
}

func QueryUpdateUser(port string, user User, uid string) *model.User {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/update.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("id", uid)
	req.Var("input", map[string]string{
		"name": user.Name,
		"type": entity.GetUserTypeString(user.Type),
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		User model.User `json:"updateUser"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return &respData.User
}

func QueryGetUser(port string, uniqueUserId string) *model.UserConnection {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/listUsers.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("filter", map[string]string{"uniqueUserId": uniqueUserId})
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
	return &respData.Users
}
