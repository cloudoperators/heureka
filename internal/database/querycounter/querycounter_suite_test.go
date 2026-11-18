// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package querycounter

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQueryCounter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "QueryCounter Suite")
}

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})
