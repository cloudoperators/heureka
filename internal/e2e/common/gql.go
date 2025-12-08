// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"context"
	"fmt"
	"os"

	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/machinebox/graphql"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var GqlStandardHeaders map[string]string = map[string]string{
	"Cache-Control": "no-cache",
}

func ExecuteGqlQueryFromFile[T any](port string, queryFilePath string, vars map[string]interface{}) T {
	return ExecuteGqlQueryFromFileWithHeaders[T](port, queryFilePath, vars, GqlStandardHeaders)
}

func ExecuteGqlQueryFromFileWithHeaders[T any](port string, queryFilePath string, vars map[string]interface{}, headers map[string]string) T {
	client, req := newClientAndRequestForGqlFileQuery(port, queryFilePath, vars, headers)

	var result T
	if err := util2.RequestWithBackoff(func() error { return client.Run(context.Background(), req, &result) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return result
}

func newClientAndRequestForGqlFileQuery(port string, queryFilePath string, vars map[string]interface{}, headers map[string]string) (*graphql.Client, *graphql.Request) {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))
	b, err := os.ReadFile(queryFilePath)
	Expect(err).To(BeNil())
	str := string(b)
	req := graphql.NewRequest(str)
	for k, v := range vars {
		req.Var(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return client, req
}
