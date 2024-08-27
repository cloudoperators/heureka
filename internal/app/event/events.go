package event

// EventName type
type EventName string

type Event interface {
	Name() EventName
}
