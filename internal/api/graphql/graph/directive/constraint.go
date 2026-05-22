// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package directive

import (
	"context"
	"fmt"
	"net/url"

	"github.com/99designs/gqlgen/graphql"
)

func Constraint(ctx context.Context, obj any, next graphql.Resolver, minLength *int, maxLength *int, format *string) (any, error) {
	val, err := next(ctx)
	if err != nil || val == nil {
		return val, err
	}

	var str string

	switch v := val.(type) {
	case string:
		str = v
	case *string:
		if v == nil {
			return val, nil
		}

		str = *v
	default:
		return val, nil
	}

	if minLength != nil && len(str) < *minLength {
		return nil, fmt.Errorf("value must be at least %d characters", *minLength)
	}

	if maxLength != nil && len(str) > *maxLength {
		return nil, fmt.Errorf("value must not exceed %d characters", *maxLength)
	}

	if format != nil && *format == "url" {
		if parsedURL, err := url.Parse(str); err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			return nil, fmt.Errorf("value must be a valid URL")
		}
	}

	return val, nil
}
