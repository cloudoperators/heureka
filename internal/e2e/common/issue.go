// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudoperators/heureka/internal/api/graphql/graph/model"
	util2 "github.com/cloudoperators/heureka/pkg/util"

	"github.com/machinebox/graphql"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

type Issue struct {
	PrimaryName string
	Description string
	Type        string
}

func QueryCreateIssue(port string, issue Issue) *model.Issue {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/create.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("input", map[string]string{
		"primaryName": issue.PrimaryName,
		"description": issue.Description,
		"type":        issue.Type,
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		Issue model.Issue `json:"createIssue"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return &respData.Issue
}

func QueryUpdateIssue(port string, issue Issue, iid string) *model.Issue {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/update.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("id", iid)
	req.Var("input", map[string]string{
		"description": issue.Description,
		"type":        issue.Type,
	})

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		Issue model.Issue `json:"updateIssue"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return &respData.Issue
}

func QueryGetIssue(port string, issuePrimaryName string) *model.IssueConnection {
	// create a queryCollection (safe to share across requests)
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))

	//@todo may need to make this more fault proof?! What if the test is executed from the root dir? does it still work?
	b, err := os.ReadFile("../api/graphql/graph/queryCollection/issue/listIssues.graphql")
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)

	req.Var("filter", map[string]string{"primaryName": issuePrimaryName})
	req.Var("first", 1)
	req.Var("after", "0")

	req.Header.Set("Cache-Control", "no-cache")
	ctx := context.Background()

	var respData struct {
		Issues model.IssueConnection `json:"Issues"`
	}
	if err := util2.RequestWithBackoff(func() error { return client.Run(ctx, req, &respData) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return &respData.Issues
}
