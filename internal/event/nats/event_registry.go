// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"context"
	"encoding/json"
	"github.com/cloudoperators/heureka/internal/database"
	"github.com/cloudoperators/heureka/internal/event"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

type eventRegistry struct {
	db            database.Database
	nc            *nats.Conn
	subscriptions map[event.EventName]*nats.Subscription
}

func (er *eventRegistry) RegisterEventHandler(subject event.EventName, handler event.EventHandler) {
	if er.subscriptions == nil {
		er.subscriptions = make(map[event.EventName]*nats.Subscription)
	}
	if er.subscriptions[subject] != nil {
		log.Debugf("Handler for event %s already registered", subject)
	} else {
		log.Infof("Registering handler for event %s", subject)

		subscription, err := er.nc.QueueSubscribe(string(subject), "heureka-core", func(msg *nats.Msg) {
			event, err := handler.Unmarshal(msg.Data)
			if err != nil {
				//todo: log error
				return
			}
			//todo: should return error
			handler.HandleEvent(er.db, event)
		})
		if err != nil {
			//todo: log error
			return
		}
		er.subscriptions[subject] = subscription
	}
}

func (er *eventRegistry) PushEvent(event event.Event) {

	b, err := json.Marshal(event)
	log.WithField("bytes", b).WithField("error", err).WithField("event", event).Infof("Pushing event %s", event.Name())
	if err != nil {
		//todo: log error
		return
	}

	err = er.nc.Publish(string(event.Name()), b)
	if err != nil {
		//todo: log error
		return
	}
}

// NewEventRegistry returns an event registry using nats as the event bus
func NewEventRegistry(db database.Database) *eventRegistry {
	//todo: move initialization outside of this
	//todo: adding options for secure & authenticated nats connection
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to nats server", err)
	}
	return &eventRegistry{
		db: db,
		nc: nc,
	}
}

// todo: this is currently only needed for channel based version, we should remove it and start running once the first event handler
func (er *eventRegistry) Run(ctx context.Context) {
	//todo: remove
	er.subscriptions = make(map[event.EventName]*nats.Subscription)
	return
}

// todo: add to interface
func (er *eventRegistry) Shutdown() {
	for _, s := range er.subscriptions {
		s.Unsubscribe()
	}
	er.nc.Close()
}
