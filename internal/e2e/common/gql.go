package e2e_common

import (
	"context"
	"os"

	util2 "github.com/cloudoperators/heureka/pkg/util"
	"github.com/machinebox/graphql"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func ExecuteGqlQueryFromFile[T any](url string, queryFilePath string, vars map[string]interface{}, headers map[string]string) T {
	client, req := newClientAndRequestForGqlFileQuery(url, queryFilePath, vars, headers)

	var result T
	if err := util2.RequestWithBackoff(func() error { return client.Run(context.Background(), req, &result) }); err != nil {
		logrus.WithError(err).WithField("request", req).Fatalln("Error while unmarshaling")
	}
	return result
}
func newClientAndRequestForGqlFileQuery(url string, queryFilePath string, vars map[string]interface{}, headers map[string]string) (*graphql.Client, *graphql.Request) {
	client := graphql.NewClient(url)
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
