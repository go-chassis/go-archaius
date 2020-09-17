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

// Package source manage all the config source and merge configs by precedence
package source

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sync"

	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/openlog"
)

//errors
var (
	ErrKeyNotExist = errors.New("key does not exist")
)

//const
const (
	//DefaultPriority gives the default priority
	DefaultPriority      = -1
	fmtInvalidKeyWithErr = "invalid key format for %s key. key registration ignored: %s"
	fmtInvalidKey        = "invalid key format for %s key"
	fmtLoadConfigFailed  = "fail to load configuration of %s source: %s"
)

// Manager manage all sources and config from them
type Manager struct {
	sourceMapMux sync.RWMutex
	Sources      map[string]ConfigSource

	configMapMux     sync.RWMutex
	ConfigurationMap map[string]string

	dispatcher *event.Dispatcher
}

// NewManager creates an object of Manager
func NewManager() *Manager {
	configMgr := new(Manager)
	configMgr.dispatcher = event.NewDispatcher()
	configMgr.Sources = make(map[string]ConfigSource)
	configMgr.ConfigurationMap = make(map[string]string)
	return configMgr
}

// Cleanup close and cleanup config manager channel
func (m *Manager) Cleanup() error {
	// cleanup all dynamic handler
	m.sourceMapMux.RLock()
	defer m.sourceMapMux.RUnlock()
	for _, s := range m.Sources {
		err := s.Cleanup()
		if err != nil {
			return err
		}
	}
	return nil
}

//Set call set of all sources
func (m *Manager) Set(k string, v interface{}) error {
	m.sourceMapMux.RLock()
	defer m.sourceMapMux.RLock()
	var err error
	for _, s := range m.Sources {
		err = s.Set(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

//Delete call Delete of all sources
func (m *Manager) Delete(k string) error {
	m.sourceMapMux.RLock()
	defer m.sourceMapMux.RUnlock()
	var err error
	for _, s := range m.Sources {
		err = s.Delete(k)
		if err != nil {
			return err
		}
	}
	return nil
}

// Unmarshal function is used in the case when user want his yaml file to be unmarshalled to structure pointer
// Unmarshal function accepts a pointer and in called function anyone can able to get the data in passed object
// Unmarshal only accepts a pointer values
// Unmarshal returns error if obj values are 0. nil and value type.
// Procedure:
//      1. Unmarshal first checks the passed object type using reflection.
//      2. Based on type Unmarshal function will check and set the values
//      ex: If type is basic types like int, string, float then it will assigb directly values,
//          If type is map, ptr and struct then it will again send for unmarshal until it find the basic type and set the values
func (m *Manager) Unmarshal(obj interface{}) error {
	rv := reflect.ValueOf(obj)
	// only pointers are accepted
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		err := errors.New("invalid object supplied")
		openlog.Error("invalid object supplied: " + err.Error())
		return err
	}

	return m.unmarshal(rv, doNotConsiderTag)
}

// AddSource adds a source to configurationManager
func (m *Manager) AddSource(source ConfigSource) error {
	if source == nil || source.GetSourceName() == "" {
		err := errors.New("nil or invalid source supplied")
		openlog.Error("nil or invalid source supplied: " + err.Error())
		return err
	}
	sourceName := source.GetSourceName()
	m.sourceMapMux.Lock()
	_, ok := m.Sources[sourceName]
	if ok {
		err := errors.New("duplicate source supplied")
		openlog.Error("duplicate source supplied: " + err.Error())
		m.sourceMapMux.Unlock()
		return err
	}

	m.Sources[sourceName] = source
	m.sourceMapMux.Unlock()

	err := m.pullSourceConfigs(sourceName)
	if err != nil {
		err = fmt.Errorf(fmtLoadConfigFailed, sourceName, err)
		openlog.Error(err.Error())
		return err
	}
	openlog.Info("invoke dynamic handler:" + source.GetSourceName())
	go source.Watch(m)

	return nil
}

func (m *Manager) pullSourceConfigs(source string) error {
	m.sourceMapMux.RLock()
	configSource, ok := m.Sources[source]
	m.sourceMapMux.RUnlock()
	if !ok {
		err := errors.New("invalid source or source not added")
		openlog.Error("invalid source or source not added: " + err.Error())
		return err
	}

	config, err := configSource.GetConfigurations()
	if config == nil || len(config) == 0 {
		if err != nil {
			openlog.Error("Get configuration by items failed: " + err.Error())
			return err
		}

		openlog.Warn(fmt.Sprintf("empty config from %s", source))
		return nil
	}

	m.updateConfigurationMap(configSource, config)

	return nil
}

// Configs returns all the key values
func (m *Manager) Configs() map[string]interface{} {
	config := make(map[string]interface{}, 0)

	m.configMapMux.RLock()
	defer m.configMapMux.RUnlock()

	for key, sourceName := range m.ConfigurationMap {
		sValue := m.configValueBySource(key, sourceName)
		if sValue == nil {
			continue
		}
		config[key] = sValue
	}

	return config
}

// AddDimensionInfo adds the dimensionInfo to the list of which configurations needs to be pulled
func (m *Manager) AddDimensionInfo(labels map[string]string) (map[string]string, error) {
	config := make(map[string]string, 0)

	err := m.addDimensionInfo(labels)
	if err != nil {
		openlog.Error(fmt.Sprintf("failed to do add dimension info %s", err))
		return config, err
	}

	return config, nil
}

// Refresh reload the configurations of a source
func (m *Manager) Refresh(sourceName string) error {
	err := m.pullSourceConfigs(sourceName)
	if err != nil {
		openlog.Error(fmt.Sprintf(fmtLoadConfigFailed, sourceName, err))
		errorMsg := "fail to load configuration of" + sourceName + " source"
		return errors.New(errorMsg)
	}
	return nil
}

func (m *Manager) configValueBySource(configKey, sourceName string) interface{} {
	m.sourceMapMux.RLock()
	source, ok := m.Sources[sourceName]
	m.sourceMapMux.RUnlock()
	if !ok {
		return nil
	}

	configValue, err := source.GetConfigurationByKey(configKey)
	if err != nil {
		// may be before getting config, Event has deleted it so get next priority config value
		nbSource := m.findNextBestSource(configKey, sourceName)
		if nbSource != nil {
			configValue, _ := nbSource.GetConfigurationByKey(configKey)
			return configValue
		}
		return nil
	}

	return configValue
}

func (m *Manager) addDimensionInfo(labels map[string]string) error {
	m.sourceMapMux.RLock()
	defer m.sourceMapMux.RUnlock()
	for _, source := range m.Sources {
		err := source.AddDimensionInfo(labels)
		if err != nil {
			return fmt.Errorf("add dimension info for source %s failed", source.GetSourceName())
		}
	}
	return nil
}

// IsKeyExist check if key exist in cache
func (m *Manager) IsKeyExist(key string) bool {
	m.configMapMux.RLock()
	defer m.configMapMux.RUnlock()

	if _, ok := m.ConfigurationMap[key]; ok {
		return true
	}

	return false
}

// GetConfig returns the value for a particular key from cache
func (m *Manager) GetConfig(key string) interface{} {
	m.configMapMux.RLock()
	sourceName, ok := m.ConfigurationMap[key]
	m.configMapMux.RUnlock()
	if !ok {
		return nil
	}
	return m.configValueBySource(key, sourceName)
}

func (m *Manager) updateConfigurationMap(source ConfigSource, configs map[string]interface{}) error {
	m.configMapMux.Lock()
	defer m.configMapMux.Unlock()
	for key := range configs {
		sourceName, ok := m.ConfigurationMap[key]
		if !ok { // if key do not exist then add source
			m.ConfigurationMap[key] = source.GetSourceName()
			continue
		}

		m.sourceMapMux.RLock()
		currentSource, ok := m.Sources[sourceName]
		m.sourceMapMux.RUnlock()
		if !ok {
			m.ConfigurationMap[key] = source.GetSourceName()
			continue
		}

		currentSrcPriority := currentSource.GetPriority()
		if currentSrcPriority > source.GetPriority() { // lesser value has high priority
			m.ConfigurationMap[key] = source.GetSourceName()
		}
	}

	return nil
}

func (m *Manager) updateConfigurationMapByDI(source ConfigSource, configs map[string]interface{}) error {
	m.configMapMux.Lock()
	defer m.configMapMux.Unlock()
	for key := range configs {
		sourceName, ok := m.ConfigurationMap[key]
		if !ok { // if key do not exist then add source
			m.ConfigurationMap[key] = source.GetSourceName()
			continue
		}

		m.sourceMapMux.RLock()
		currentSource, ok := m.Sources[sourceName]
		m.sourceMapMux.RUnlock()
		if !ok {
			m.ConfigurationMap[key] = source.GetSourceName()
			continue
		}

		currentSrcPriority := currentSource.GetPriority()
		if currentSrcPriority > source.GetPriority() { // lesser value has high priority
			m.ConfigurationMap[key] = source.GetSourceName()
		}
	}

	return nil
}

func (m *Manager) updateModuleEvent(es []*event.Event) error {
	if es == nil || len(es) == 0 {
		return errors.New("nil or invalid events supplied")
	}

	for i := 0; i < len(es); i++ {
		m.updateEvent(es[i])
	}

	return m.dispatcher.DispatchModuleEvent(es)
}

func (m *Manager) updateEvent(e *event.Event) error {
	// refresh all configuration one by one
	if e == nil || e.EventSource == "" || e.Key == "" {
		return errors.New("nil or invalid event supplied")
	}
	openlog.Info("config update event received")
	switch e.EventType {
	case event.Create, event.Update:
		m.configMapMux.RLock()
		sourceName, ok := m.ConfigurationMap[e.Key]
		m.configMapMux.RUnlock()
		if !ok {
			m.configMapMux.Lock()
			m.ConfigurationMap[e.Key] = e.EventSource
			m.configMapMux.Unlock()
			e.EventType = event.Create
		} else if sourceName == e.EventSource {
			e.EventType = event.Update
		} else if sourceName != e.EventSource {
			prioritySrc := m.getHighPrioritySource(sourceName, e.EventSource)
			if prioritySrc != nil && prioritySrc.GetSourceName() == sourceName {
				// if event generated from less priority source then ignore
				openlog.Info(fmt.Sprintf("the event source %s's priority is less then %s's, ignore",
					e.EventSource, sourceName))
				return nil
			}
			m.configMapMux.Lock()
			m.ConfigurationMap[e.Key] = e.EventSource
			m.configMapMux.Unlock()
			e.EventType = event.Update
		}

	case event.Delete:
		m.configMapMux.RLock()
		sourceName, ok := m.ConfigurationMap[e.Key]
		m.configMapMux.RUnlock()
		if !ok || sourceName != e.EventSource {
			// if delete event generated from source not maintained ignore it
			return nil
		} else if sourceName == e.EventSource {
			// find less priority source or delete key
			source := m.findNextBestSource(e.Key, sourceName)
			if source == nil {
				m.configMapMux.Lock()
				delete(m.ConfigurationMap, e.Key)
				m.configMapMux.Unlock()
			} else {
				m.configMapMux.Lock()
				m.ConfigurationMap[e.Key] = source.GetSourceName()
				m.configMapMux.Unlock()
			}
		}

	}

	m.dispatcher.DispatchEvent(e)

	return nil
}

// OnEvent Triggers actions when an event is generated
func (m *Manager) OnEvent(event *event.Event) {
	err := m.updateEvent(event)
	if err != nil {
		openlog.Error("failed in updating event with error: " + err.Error())
	}
}

// OnModuleEvent Triggers actions when events are generated
func (m *Manager) OnModuleEvent(event []*event.Event) {
	if err := m.updateModuleEvent(event); err != nil {
		openlog.Error("failed in updating events with error: " + err.Error())
	}
}

func (m *Manager) findNextBestSource(key string, sourceName string) ConfigSource {
	var rSource ConfigSource
	m.sourceMapMux.RLock()
	for _, source := range m.Sources {
		if source.GetSourceName() == sourceName {
			continue
		}
		value, err := source.GetConfigurationByKey(key)
		if err != nil || value == nil {
			continue
		}
		if rSource == nil {
			rSource = source
			continue
		}
		if source.GetPriority() < rSource.GetPriority() { // less value has high priority
			rSource = source
		}
	}
	m.sourceMapMux.RUnlock()

	return rSource
}

func (m *Manager) getHighPrioritySource(srcNameA, srcNameB string) ConfigSource {
	m.sourceMapMux.RLock()
	sourceA, okA := m.Sources[srcNameA]
	sourceB, okB := m.Sources[srcNameB]
	m.sourceMapMux.RUnlock()

	if !okA && !okB {
		return nil
	} else if !okA {
		return sourceB
	} else if !okB {
		return sourceA
	}

	if sourceA.GetPriority() < sourceB.GetPriority() { //less value has high priority
		return sourceA
	}

	return sourceB
}

// RegisterListener Function to Register all listener for different key changes
func (m *Manager) RegisterListener(listenerObj event.Listener, keys ...string) error {
	for _, key := range keys {
		_, err := regexp.Compile(key)
		if err != nil {
			openlog.Error(fmt.Sprintf(fmtInvalidKeyWithErr, key, err))
			return fmt.Errorf(fmtInvalidKey, key)
		}
	}

	return m.dispatcher.RegisterListener(listenerObj, keys...)
}

// UnRegisterListener remove listener
func (m *Manager) UnRegisterListener(listenerObj event.Listener, keys ...string) error {
	for _, key := range keys {
		_, err := regexp.Compile(key)
		if err != nil {
			openlog.Error(fmt.Sprintf(fmtInvalidKeyWithErr, key, err))
			return fmt.Errorf(fmtInvalidKey, key)
		}
	}

	return m.dispatcher.UnRegisterListener(listenerObj, keys...)
}

// RegisterModuleListener Function to Register all moduleListener for different key(prefix) changes
func (m *Manager) RegisterModuleListener(listenerObj event.ModuleListener, prefixes ...string) error {
	for _, prefix := range prefixes {
		if prefix == "" {
			openlog.Error(fmt.Sprintf(fmtInvalidKey, prefix))
			return fmt.Errorf(fmtInvalidKey, prefix)
		}
	}

	return m.dispatcher.RegisterModuleListener(listenerObj, prefixes...)
}

// UnRegisterModuleListener remove moduleListener
func (m *Manager) UnRegisterModuleListener(listenerObj event.ModuleListener, prefixes ...string) error {
	for _, prefix := range prefixes {
		if prefix == "" {
			openlog.Error(fmt.Sprintf(fmtInvalidKey, prefix))
			return fmt.Errorf(fmtInvalidKey, prefix)
		}
	}

	return m.dispatcher.UnRegisterModuleListener(listenerObj, prefixes...)
}
