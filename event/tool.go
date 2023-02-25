/*
 * Copyright 2020 Huawei Technologies Co., Ltd
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

package event

import (
	"reflect"
)

// PopulateEvents compare old and new configurations to generate events that need to be updated
func PopulateEvents(sourceName string, currentConfig, updatedConfig map[string]interface{}) ([]*Event, error) {
	events := make([]*Event, 0)

	// generate create and update event
	for key, value := range updatedConfig {
		if currentConfig != nil {
			currentValue, ok := currentConfig[key]
			if !ok { // if new configuration introduced
				events = append(events, constructEvent(sourceName, Create, key, value))
			} else if !reflect.DeepEqual(currentValue, value) {
				events = append(events, constructEvent(sourceName, Update, key, value))
			}
		} else {
			events = append(events, constructEvent(sourceName, Create, key, value))
		}

	}

	// generate delete event
	for key, value := range currentConfig {
		_, ok := updatedConfig[key]
		if !ok { // when old config not present in new config
			events = append(events, constructEvent(sourceName, Delete, key, value))
		}
	}
	return events, nil
}

func constructEvent(sourceName string, eventType string, key string, value interface{}) *Event {
	newEvent := new(Event)
	newEvent.EventSource = sourceName
	newEvent.EventType = eventType
	newEvent.Key = key
	newEvent.Value = value

	return newEvent
}
