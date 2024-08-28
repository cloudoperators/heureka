// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

// EventName type
type EventName string

type Event interface {
	Name() EventName
}
