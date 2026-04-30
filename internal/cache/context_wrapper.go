// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package cache

import "context"

func WrapContext2[A1 any, A2 any, R any](ctx context.Context, f func(context.Context, A1, A2) (R, error)) func(A1, A2) (R, error) {
	return func(a1 A1, a2 A2) (R, error) { return f(ctx, a1, a2) }
}
