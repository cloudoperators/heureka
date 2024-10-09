// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	"github.com/cloudoperators/heureka/internal/entity"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

type User struct {
	Id   string
	Type entity.UserType
	Name string
}

func QueryCreateUser(port string, user User) *model.User {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/create.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("input", map[string]string{
		"uniqueUserId": user.Id,
		"type":         entity.GetUserTypeString(user.Type),
		"name":         user.Name,
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

func QueryUpdateUser(port string, user User) *model.User {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/update.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("id", user.Id)
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

func QueryGetUser(port string, userId string) *model.UserConnection {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/user/listUsers.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("filter", map[string]string{"uniqueUserId": userId})
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
