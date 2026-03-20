// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_common

import (
	"context"
	"fmt"
	"os"

	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/machinebox/graphql"
)

func GetRandomFreePort() string {
	return util2.GetRandomFreePort()
}

var GqlStandardHeaders map[string]string = map[string]string{
	"Cache-Control": "no-cache",
}

func ExecuteGqlQuery[T any](port string, query string, vars map[string]any) (T, error) {
	return ExecuteGqlQueryWithHeaders[T](port, query, vars, nil)
}

func ExecuteGqlQueryWithHeaders[T any](
	port string,
	query string,
	vars map[string]any,
	headers map[string]string,
) (T, error) {
	client := graphql.NewClient(fmt.Sprintf("http://localhost:%s/query", port))
	req := graphql.NewRequest(query)

	for k, v := range vars {
		req.Var(k, v)
	}

	for k, v := range GqlStandardHeaders {
		req.Header.Set(k, v)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return executeGql[T](client, req)
}

func executeGql[T any](client *graphql.Client, req *graphql.Request) (T, error) {
	var result T

	err := util2.RequestWithBackoff(func() error {
		return client.Run(context.Background(), req, &result)
	})

	return result, err
}

func ExecuteGqlQueryFromFile[T any](port string, queryFilePath string, vars map[string]any) (T, error) {
	return ExecuteGqlQueryFromFileWithHeaders[T](port, queryFilePath, vars, GqlStandardHeaders)
}

func ExecuteGqlQueryFromFileWithHeaders[T any](port string, queryFilePath string, vars map[string]any,
	headers map[string]string,
) (T, error) {
	b, err := os.ReadFile(queryFilePath)
	if err != nil {
		var zero T
		return zero, err
	}

	return ExecuteGqlQueryWithHeaders[T](
		port,
		string(b),
		vars,
		headers,
	)
}
