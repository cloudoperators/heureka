// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package graphqlapi

import (
	"embed"
	"strings"
)

//go:embed graph/schema/*.graphqls
var schemaFiles embed.FS

func Schema() string {
	entries, err := schemaFiles.ReadDir("graph/schema")
	if err != nil {
		panic("failed to read embedded schema directory: " + err.Error())
	}

	var sb strings.Builder

	for _, entry := range entries {
		data, err := schemaFiles.ReadFile("graph/schema/" + entry.Name())
		if err != nil {
			panic("failed to read embedded schema file " + entry.Name() + ": " + err.Error())
		}

		sb.Write(data)
		sb.WriteByte('\n')
	}

	return sb.String()
}
