// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package graph

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cloudoperators/heureka/internal/api/graphql/graph/baseResolver"
)

// ErrorPresenter is a custom error presenter for gqlgen
// This converts our GraphQLError to the format expected by gqlgen
func ErrorPresenter(ctx context.Context, e error) *graphql.Error {
	// If it's already our custom GraphQL error, use it directly
	var gqlErr *baseResolver.GraphQLError
	if errors.As(e, &gqlErr) {
		return &graphql.Error{
			Message:    gqlErr.Message,
			Path:       gqlErr.Path,
			Extensions: gqlErr.Extensions,
		}
	}

	// Otherwise, convert the error to our GraphQL error format
	customErr := baseResolver.ToGraphQLError(e)

	// Then convert to gqlgen's graphql.Error
	return &graphql.Error{
		Message:    customErr.Message,
		Extensions: customErr.Extensions,
	}
}
