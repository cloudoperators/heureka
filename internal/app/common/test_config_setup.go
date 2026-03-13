// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"os"

	"github.com/cloudoperators/heureka/internal/util"
)

func GetTestConfig(authEnabled bool) *util.Config {
	var modelFilePath = "./../../openfga/model/model.fga"

	cfg := &util.Config{
		CurrentUser: "testuser",
	}

	if authEnabled {
		cfg.AuthzOpenFgaApiUrl = os.Getenv("AUTHZ_FGA_API_URL")
		cfg.AuthzOpenFgaApiToken = os.Getenv("AUTHZ_FGA_API_TOKEN")
		cfg.AuthzOpenFgaStoreName = os.Getenv("AUTHZ_FGA_STORE_NAME")
		cfg.AuthzModelFilePath = modelFilePath
	}

	return cfg
}
