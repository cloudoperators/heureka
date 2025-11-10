// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package common

import "github.com/cloudoperators/heureka/internal/entity"

func EnsureListOptions(options *entity.ListOptions) *entity.ListOptions {
	if options == nil {
		options = &entity.ListOptions{
			Order: []entity.Order{},
		}
	}
	return options
}
