// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"

	"github.wdf.sap.corp/cc/heureka/scanners/nvd/processor"
)

var cveProcessor *processor.Processor
var cfg processor.Config

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "e2e Suite")
}

var _ = BeforeSuite(func() {
	cfg = processor.Config {
		HeurekaUrl: "http://127.0.0.1:80/query",
		IssueRepositoryName: "NVD",
		IssueRepositoryUrl: "https://nvd.nist.gov",
	}

	// Setup new processor
	cveProcessor = processor.NewProcessor(cfg)
	cveProcessor.IssueRepositoryName = cfg.IssueRepositoryName
})
