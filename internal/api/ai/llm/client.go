// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package llm

import "context"

type Client interface {
	Complete(ctx context.Context, systemPrompt, userMessage string) (string, error)
}
