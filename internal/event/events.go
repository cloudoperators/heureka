// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0
package event

import (
	"github.com/cloudoperators/heureka/internal/database"
)

// EventName type
type EventName string

type Event interface {
	Name() EventName
}

type EventHandlerFunc func(database.Database, Event)

func (f EventHandlerFunc) HandleEvent(db database.Database, e Event) {
	f(db, e)
}

type EventHandler struct {
	Handler     func(database.Database, Event)
	Unmarshaler func(data []byte) (Event, error)
}

func (h *EventHandler) HandleEvent(db database.Database, e Event) {
	h.Handler(db, e)
}

func (h *EventHandler) Unmarshal(data []byte) (Event, error) {
	return h.Unmarshaler(data)
}
