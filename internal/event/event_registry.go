// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"github.com/cloudoperators/heureka/internal/database"
)

type eventRegistry struct {
	handlers map[EventName][]EventHandler
	db       database.Database
	ch       chan Event
}

func (er *eventRegistry) RegisterEventHandler(event EventName, handler EventHandler) {
	if er.handlers == nil {
		er.handlers = make(map[EventName][]EventHandler)
	}

	er.handlers[event] = append(er.handlers[event], handler)
}

func (er *eventRegistry) PushEvent(event Event) {
	if er.ch == nil {
		er.ch = make(chan Event, 1000)
	}

	er.ch <- event
}

// NewEventRegistry returns an event registry where for each incoming event a list of
// handlers is called. We use a buffered channel for the worker go routines.
func NewEventRegistry(db database.Database) *eventRegistry {
	return &eventRegistry{
		handlers: make(map[EventName][]EventHandler),
		ch:       make(chan Event, 1000),
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
				handler.HandleEvent(er.db, event)
			}
		}
	}
}
