/*
 * Copyright 2017 Huawei Technologies Co., Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * Created by on 2017/6/22.
 */

//Package event provides the different Listeners
package event

import (
	"errors"
	"github.com/go-chassis/openlog"
	"regexp"
	"strings"
)

//errors
var (
	ErrNilListener = errors.New("nil listener")
)

// Event Constant
const (
	Update        = "UPDATE"
	Delete        = "DELETE"
	Create        = "CREATE"
	InvalidAction = "INVALID-ACTION"
)

// Event generated when any config changes
type Event struct {
	EventSource string
	EventType   string
	Key         string
	Value       interface{}
}

// Listener All Listener should implement this Interface
type Listener interface {
	Event(event *Event)
}

// ModuleListener All moduleListener should implement this Interface
type ModuleListener interface {
	Event(event []*Event)
}

//Dispatcher is the observer
type Dispatcher struct {
	listeners       map[string][]Listener
	moduleListeners map[string][]ModuleListener
}

// NewDispatcher is a new Dispatcher for listeners
func NewDispatcher() *Dispatcher {
	dis := new(Dispatcher)
	dis.listeners = make(map[string][]Listener)
	dis.moduleListeners = make(map[string][]ModuleListener)
	return dis
}

// RegisterListener registers listener for particular configuration
func (dis *Dispatcher) RegisterListener(listenerObj Listener, keys ...string) error {
	if listenerObj == nil {
		err := ErrNilListener
		openlog.Error("nil listener supplied:" + err.Error())
		return ErrNilListener
	}

	for _, key := range keys {
		listenerList, ok := dis.listeners[key]
		if !ok {
			listenerList = make([]Listener, 0)
		}

		// for duplicate registration
		for _, listener := range listenerList {
			if listener == listenerObj {
				return nil
			}
		}

		// append new listener
		listenerList = append(listenerList, listenerObj)

		// assign latest listener list
		dis.listeners[key] = listenerList
	}
	return nil
}

// UnRegisterListener un-register listener for a particular configuration
func (dis *Dispatcher) UnRegisterListener(listenerObj Listener, keys ...string) error {
	if listenerObj == nil {
		return ErrNilListener
	}

	for _, key := range keys {
		listenerList, ok := dis.listeners[key]
		if !ok {
			continue
		}

		newListenerList := make([]Listener, 0)
		// remove listener
		for _, listener := range listenerList {
			if listener == listenerObj {
				continue
			}
			newListenerList = append(newListenerList, listener)
		}

		// assign latest listener list
		dis.listeners[key] = newListenerList
	}
	return nil
}

// DispatchEvent sends the action trigger for a particular event on a configuration
func (dis *Dispatcher) DispatchEvent(event *Event) error {
	if event == nil {
		return errors.New("empty event provided")
	}

	for regKey, listeners := range dis.listeners {
		matched, err := regexp.MatchString(regKey, event.Key)
		if err != nil {
			openlog.Error("regular expression for key " + regKey + " failed:" + err.Error())
			continue
		}
		if matched {
			for _, listener := range listeners {
				openlog.Info("event generated for " + regKey)
				go listener.Event(event)
			}
		}
	}

	return nil
}

// RegisterModuleListener registers moduleListener for particular configuration
func (dis *Dispatcher) RegisterModuleListener(listenerObj ModuleListener, modulePrefixes ...string) error {
	if listenerObj == nil {
		err := ErrNilListener
		openlog.Error("nil moduleListener supplied:" + err.Error())
		return ErrNilListener
	}

	for _, prefix := range modulePrefixes {
		moduleListeners, ok := dis.moduleListeners[prefix]
		if !ok {
			moduleListeners = make([]ModuleListener, 0)
		}

		// for duplicate registration
		for _, listener := range moduleListeners {
			if listener == listenerObj {
				return nil
			}
		}

		// append new moduleListener
		moduleListeners = append(moduleListeners, listenerObj)

		// assign latest moduleListener list
		dis.moduleListeners[prefix] = moduleListeners
	}
	return nil
}

// UnRegisterModuleListener un-register moduleListener for a particular configuration
func (dis *Dispatcher) UnRegisterModuleListener(listenerObj ModuleListener, modulePrefixes ...string) error {
	if listenerObj == nil {
		return ErrNilListener
	}

	for _, prefix := range modulePrefixes {
		listenerList, ok := dis.moduleListeners[prefix]
		if !ok {
			continue
		}

		newListenerList := make([]ModuleListener, 0)
		// remove moduleListener
		for _, listener := range listenerList {
			if listener == listenerObj {
				continue
			}
			newListenerList = append(newListenerList, listener)
		}

		// assign latest moduleListener list
		dis.moduleListeners[prefix] = newListenerList
	}
	return nil
}

// DispatchModuleEvent finds the registered function for callback according to the prefix of key in events
func (dis *Dispatcher) DispatchModuleEvent(events []*Event) error {
	if events == nil || len(events) == 0 {
		return errors.New("empty events provided")
	}

	// 1. According to the key in the event, events with the same prefix are placed in the same slice
	eventsList := dis.parseEvents(events)

	// 2. Events with the same prefix will only be callback once.
	for key, events := range eventsList {
		if listeners, ok := dis.moduleListeners[key]; ok {
			for _, listener := range listeners {
				openlog.Info("events generated for " + key)
				go listener.Event(events)
			}
		}
	}

	return nil
}

// Event key with the same subscription prefix is placed in the same slice
func (dis *Dispatcher) parseEvents(events []*Event) map[string][]*Event {
	var eventList = make(map[string][]*Event)
	for _, event := range events {
		// find first prefix from event.key
		registerKey := dis.findFirstRegisterPrefix(event.Key)
		if module, ok := eventList[registerKey]; ok {
			events := module
			events = append(events, event)
			eventList[registerKey] = events
		} else {
			newModule := append([]*Event{}, event)
			eventList[registerKey] = newModule
		}
	}

	return eventList
}

// Find first prefix from event.key
// Ignore the case where namespace and module key(prefix) have the same name
func (dis *Dispatcher) findFirstRegisterPrefix(eventKey string) string {
	keyArr := strings.Split(eventKey, ".")
	for _, key := range keyArr {
		if _, ok := dis.moduleListeners[key]; ok {
			return key
		}
	}
	return ""
}
