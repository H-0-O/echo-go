package util

import (
	"strings"
)

// EventFormatter helps format event names with namespaces.
type EventFormatter struct {
	Namespace string
}

// NewEventFormatter creates a new EventFormatter.
func NewEventFormatter(namespace string) *EventFormatter {
	return &EventFormatter{
		Namespace: namespace,
	}
}

// Format an event name.
func (f *EventFormatter) Format(event string) string {
	if strings.HasPrefix(event, ".") || strings.HasPrefix(event, "\\") {
		return event[1:]
	}

	if f.Namespace != "" {
		return f.Namespace + "." + event
	}

	return event
}

// SetNamespace sets the namespace.
func (f *EventFormatter) SetNamespace(value string) {
	f.Namespace = value
}
