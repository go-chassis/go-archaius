package core

import (
	"regexp"
	"unsafe"
)

const (
	UPDATE = "UPDATE"
	DELETE = "DELETE"
	CREATE = "CREATE"
)

type Dispatcher struct {
	listeners map[string]*EventChain
}

type EventChain struct {
	callbacks []*EventCallback
}

func createEventChain() *EventChain {
	return &EventChain{callbacks: []*EventCallback{}}
}

type Event struct {
	EventSource string
	EventName   string //This is the configuration key when event happened
	EventType   string
	Value       interface{}
}

func CreateEvent(eventSource string, eventName string, eventType string, params interface{}) *Event {
	return &Event{EventSource: eventSource, EventName: eventName, EventType: eventType, Value: params}
}

type EventCallback func(*Event)

var instance *Dispatcher

func DefaultDispatcher() *Dispatcher {
	if instance == nil {
		instance = &Dispatcher{}
		instance.Init()
	}

	return instance
}

func NewDispatcher() *Dispatcher {
	i := &Dispatcher{}
	i.Init()

	return i
}

func (this *Dispatcher) Init() {
	this.listeners = make(map[string]*EventChain)
}

func (this *Dispatcher) AddEventListener(eventName string, callback *EventCallback) {
	eventChain, ok := this.listeners[eventName]
	if !ok {
		eventChain = createEventChain()
		this.listeners[eventName] = eventChain
	}

	exist := false
	for _, item := range eventChain.callbacks {
		a := *(*int)(unsafe.Pointer(item))
		b := *(*int)(unsafe.Pointer(callback))
		if a == b {
			exist = true
			break
		}
	}

	if exist {
		return
	}

	eventChain.callbacks = append(eventChain.callbacks[:], callback)
	return
}

func (this *Dispatcher) RemoveEventListener(eventName string, callback *EventCallback) {
	eventChain, ok := this.listeners[eventName]
	if !ok {
		return
	}

	exist := false
	key := 0
	for k, item := range eventChain.callbacks {
		a := *(*int)(unsafe.Pointer(item))
		b := *(*int)(unsafe.Pointer(callback))
		if a == b {
			exist = true
			key = k
			break
		}
	}

	if exist {
		eventChain.callbacks = append(eventChain.callbacks[:key], eventChain.callbacks[key+1:]...)
	}
}

func (this *Dispatcher) DispatchEvent(event *Event) {
	for key, item := range this.listeners {
		matched, _ := regexp.MatchString(key, event.EventName)
		if matched {
			for _, callback := range item.callbacks {
				go (*callback)(event)
			}
		}
	}
}
