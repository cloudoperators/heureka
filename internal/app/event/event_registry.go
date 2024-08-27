// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package event

import (
	"context"
	"sync"
)

type Handler func(Event)

type EventRegistry interface {
	RegisterEventHandler(EventName, Handler)
	PushEvent(Event)
	Run(ctx context.Context)
}

type eventRegistry struct {
	events   []Event
	handlers map[EventName][]Handler
	mu       sync.Mutex
}

func (er *eventRegistry) RegisterEventHandler(event EventName, handler Handler) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.handlers == nil {
		er.handlers = make(map[EventName][]Handler)
	}

	er.handlers[event] = append(er.handlers[event], handler)
}

func (er *eventRegistry) PushEvent(event Event) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if er.events == nil {
		er.events = make([]Event, 0)
	}

	er.events = append(er.events, event)
}

func NewEventRegistry() EventRegistry {
	return &eventRegistry{}
}

func (er *eventRegistry) Run(ctx context.Context) {
	go er.process(ctx)
}

func (er *eventRegistry) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if len(er.events) == 0 {
				continue
			}

			er.mu.Lock()
			event := er.events[0]
			er.events = er.events[1:]
			er.mu.Unlock()

			for _, handler := range er.handlers[event.Name()] {
				handler(event)
			}
		}
	}
}
