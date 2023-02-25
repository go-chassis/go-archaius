package event_test

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	"github.com/arielsrv/go-archaius/event"
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

type MListener struct {
	eventKeys []string
	wg        sync.WaitGroup
}

func (m *MListener) Event(events []*event.Event) {
	for _, ev := range events {
		m.eventKeys = append(m.eventKeys, ev.Key)
		m.wg.Done()
	}
	m.wg.Done()
}

func TestDispatcher_DispatchModuleEvent(t *testing.T) {
	t.Run("RegisterModuleEvent", func(t *testing.T) {
		dispatcher := event.NewDispatcher()
		lis := &MListener{}
		dispatcher.RegisterModuleListener(lis, "aaa.bbb")
		lis.wg.Add(3)
		dispatcher.DispatchModuleEvent([]*event.Event{
			{
				Key: "aaa.bbb.ccc",
			},
			{
				Key: "aaa",
			},
			{
				Key: "aaa.bbb",
			},
		})
		lis.wg.Wait()
		if assert.Len(t, lis.eventKeys, 2) {
			assert.Equal(t, "aaa.bbb.ccc", lis.eventKeys[0])
			assert.Equal(t, "aaa.bbb", lis.eventKeys[1])
		}
	})
	t.Run("RegisterModuleEventCovered", func(t *testing.T) {
		dispatcher := event.NewDispatcher()
		lis1 := &MListener{}
		dispatcher.RegisterModuleListener(lis1, "aaa.bbb")
		lis2 := &MListener{}
		dispatcher.RegisterModuleListener(lis2, "aaa.bbb.ccc")
		lis1.wg.Add(3)
		dispatcher.DispatchModuleEvent([]*event.Event{
			{
				Key: "aaa.bbb.ccc",
			},
			{
				Key: "aaa",
			},
			{
				Key: "aaa.bbb",
			},
		})
		lis1.wg.Wait()
		if assert.Len(t, lis1.eventKeys, 2) {
			assert.Equal(t, "aaa.bbb.ccc", lis1.eventKeys[0])
			assert.Equal(t, "aaa.bbb", lis1.eventKeys[1])
		}
		assert.Len(t, lis2.eventKeys, 0)
	})
	t.Run("UnRegisterModuleEventCovered", func(t *testing.T) {
		dispatcher := event.NewDispatcher()
		lis1 := &MListener{}
		dispatcher.RegisterModuleListener(lis1, "aaa.bbb")
		lis2 := &MListener{}
		dispatcher.RegisterModuleListener(lis2, "aaa.bbb.ccc")
		dispatcher.UnRegisterModuleListener(lis1, "aaa.bbb")
		lis2.wg.Add(2)
		dispatcher.DispatchModuleEvent([]*event.Event{
			{
				Key: "aaa.bbb.ccc",
			},
			{
				Key: "aaa",
			},
			{
				Key: "aaa.bbb",
			},
		})
		lis2.wg.Wait()
		if assert.Len(t, lis2.eventKeys, 1) {
			assert.Equal(t, "aaa.bbb.ccc", lis2.eventKeys[0])
		}
	})
}
