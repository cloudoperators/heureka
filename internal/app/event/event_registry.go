// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"github.wdf.sap.corp/cc/heureka/internal/database"
)

type Handler func(database.Database, Event)

type EventRegistry interface {
	RegisterEventHandler(EventName, Handler)
	PushEvent(Event)
	Run(ctx context.Context)
}

type eventRegistry struct {
	handlers map[EventName][]Handler
	db       database.Database
	ch       chan Event
}

func (er *eventRegistry) RegisterEventHandler(event EventName, handler Handler) {
	if er.handlers == nil {
		er.handlers = make(map[EventName][]Handler)
	}

	er.handlers[event] = append(er.handlers[event], handler)
}

func (er *eventRegistry) PushEvent(event Event) {
	if er.ch == nil {
		er.ch = make(chan Event, 100)
	}

	er.ch <- event
}

func NewEventRegistry(db database.Database) EventRegistry {
	return &eventRegistry{
		handlers: make(map[EventName][]Handler),
		ch:       make(chan Event, 100),
		db:       db,
	}
}

func (er *eventRegistry) Run(ctx context.Context) {
	go er.process(ctx)
}

func (er *eventRegistry) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-er.ch:
			for _, handler := range er.handlers[event.Name()] {
				handler(er.db, event)
			}
		}
	}
}
