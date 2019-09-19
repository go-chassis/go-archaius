package event_test

import (
	"github.com/go-chassis/go-archaius/event"
	"testing"
)

type EListener struct {
	Name      string
	EventName string
}

func (e *EListener) Event(event *event.Event) {
	e.EventName = event.Key
}

func TestDispatchEvent(t *testing.T) {
	dispatcher := event.NewDispatcher()
	var e *event.Event
	err := dispatcher.DispatchEvent(e)
	if err == nil {
		t.Error("Dispatcher failed to identify the nil event")
	}

	eventListener1 := &EListener{Name: "eventListener"}
	eventListener2 := &EListener{Name: "eventListener"}
	eventListener3 := &EListener{Name: "eventListener"}
	err = dispatcher.RegisterListener(eventListener1, "*")

	e = &event.Event{Key: "TestKey", Value: "TestValue"}
	err = dispatcher.DispatchEvent(e)
	if err != nil {
		t.Error("dispatches the event for regular expresssion failed key")
	}

	dispatcher.RegisterListener(eventListener2, "Key1")
	dispatcher.RegisterListener(eventListener3, "Key1")

	//unregister

	var listener event.Listener = nil
	//supplying nil listener
	err = dispatcher.UnRegisterListener(listener, "key")
	if err == nil {
		t.Error("event system processing on nil listener")
	}

	err = dispatcher.UnRegisterListener(eventListener1, "unregisteredkey")
	if err != nil {
		t.Error("event system unable to identify the unregisteredkey")
	}

	err = dispatcher.UnRegisterListener(eventListener2, "Key1")
	if err != nil {
		t.Error("event system unable to identify the unregisteredkey")
	}

	dispatcher.UnRegisterListener(eventListener3, "Key1")
	dispatcher.UnRegisterListener(eventListener1, "*")

	//register

	t.Log("supplying nil listener")
	err = dispatcher.RegisterListener(listener, "key")
	if err == nil {
		t.Error("Event system working on nil listener")
	}

	err = dispatcher.RegisterListener(eventListener3, "Key1")
	if err != nil {
		t.Error("Event system working on nil listener")
	}

	t.Log("duplicate registration")
	err = dispatcher.RegisterListener(eventListener3, "Key1")
	if err != nil {
		t.Error("Failed to detect the duplicate registration")
	}

	dispatcher.UnRegisterListener(eventListener3, "Key1")

}
