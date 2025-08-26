// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package openfga_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudoperators/heureka/internal/openfga"
	"github.com/cloudoperators/heureka/internal/util"
)

var _ = Describe("Authz", func() {
	var (
		cfg   *util.Config
		authz openfga.Authorization
	)

	BeforeEach(func() {
		enableLogs := true
		cfg = &util.Config{
			AuthzEnabled: true,
		}
		authz = openfga.NewAuthorizationHandler(cfg, enableLogs)
	})

	Describe("NewAuthz", func() {
		It("should create a new Authz instance", func() {
			Expect(authz).NotTo(BeNil())
		})
	})
})
