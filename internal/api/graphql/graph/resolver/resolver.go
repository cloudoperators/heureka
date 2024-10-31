// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package resolver

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/cloudoperators/heureka/internal/api/graphql/graph"
	"github.com/cloudoperators/heureka/internal/app"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	App app.Heureka
}

func NewResolver(a app.Heureka) graph.Config {

	r := Resolver{
		App: a,
	}

	return graph.Config{
		Resolvers: &r,
	}
}
