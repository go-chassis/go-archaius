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
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
	"regexp"
	"sync"

	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/openlog"
)

//errors
var (
	ErrKeyNotExist   = errors.New("key does not exist")
	ErrIgnoreChange  = errors.New("ignore key changed")
	ErrWriterInvalid = errors.New("writer is invalid")
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

	configUpdateMux sync.Mutex
	ConfigValueCache sync.Map
	ConfigurationMap sync.Map

	dispatcher *event.Dispatcher
}

// NewManager creates an object of Manager
func NewManager() *Manager {
	configMgr := new(Manager)
	configMgr.dispatcher = event.NewDispatcher()
	configMgr.Sources = make(map[string]ConfigSource)
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

	m.ConfigValueCache = sync.Map{}
	
	return nil
}

//Set call set of all sources
func (m *Manager) Set(k string, v interface{}) error {
	m.sourceMapMux.RLock()
	defer m.sourceMapMux.RUnlock()
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

// Marshal function is used to write all configuration by yaml
func (m *Manager) Marshal(w io.Writer) error {
	if w == nil {
		openlog.Error("invalid writer")
		return ErrWriterInvalid
	}
	allConfig := make(map[string]map[string]interface{})
	for name, source := range m.Sources {
		config, err := source.GetConfigurations()
		if err != nil {
			openlog.Error("get source " + name + " error " + err.Error())
			continue
		}
		if len(config) == 0 {
			continue
		}
		allConfig[name] = config
	}
	encode := yaml.NewEncoder(w)
	return encode.Encode(allConfig)
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

	m.updateConfigurationMapWithConfigUpdateLock(configSource, config)

	return nil
}

// Configs returns all the key values
func (m *Manager) Configs() map[string]interface{} {
	config := make(map[string]interface{}, 0)

	m.ConfigValueCache.Range(func(key, value interface{}) bool {
		if value == nil {
			return true
		}
		config[key.(string)] = value
		return true
	})

	return config
}

// ConfigsWithSourceNames returns all the key values along with its source name
// the returned map will be like:
// map[string]interface{}{
// 		key string: map[string]interface{"value": value, "source": sourceName}
// }
func (m *Manager) ConfigsWithSourceNames() map[string]interface{} {
	config := make(map[string]interface{}, 0)

	m.ConfigurationMap.Range(func(key, value interface{}) bool {
		sValue, ok := m.ConfigValueCache.Load(key.(string))
		if !ok || sValue == nil {
			return true
		}
		// each key stores its value and source name
		config[key.(string)] = map[string]interface{}{"value": sValue, "source": value}
		return true
	})
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

	if _, ok := m.ConfigurationMap.Load(key); ok {
		return true
	}

	return false
}

// GetConfig returns the value for a particular key from cache
func (m *Manager) GetConfig(key string) interface{} {
	val, _ := m.ConfigValueCache.Load(key)
	return val
}

func (m *Manager) updateConfigurationMapWithConfigUpdateLock(source ConfigSource, configs map[string]interface{}) error {
	m.configUpdateMux.Lock()
	defer m.configUpdateMux.Unlock()
	for key := range configs {

		val, err := source.GetConfigurationByKey(key)
		if err != nil {
			return err
		}

		sourceName, ok := m.ConfigurationMap.Load(key)
		if !ok { // if key do not exist then add source
			m.updateCacheWithoutConfigUpdateLock(source.GetSourceName(), key, val)
			continue
		}

		m.sourceMapMux.RLock()
		currentSource, ok := m.Sources[sourceName.(string)]
		m.sourceMapMux.RUnlock()
		if !ok {
			m.updateCacheWithoutConfigUpdateLock(source.GetSourceName(), key, val)
			continue
		}

		currentSrcPriority := currentSource.GetPriority()
		if currentSrcPriority > source.GetPriority() { // lesser value has high priority
			m.updateCacheWithoutConfigUpdateLock(source.GetSourceName(), key, val)
		}
	}

	return nil
}

func (m *Manager) updateModuleEvent(es []*event.Event) error {
	if es == nil || len(es) == 0 {
		return errors.New("nil or invalid events supplied")
	}

	var validEvents []*event.Event
	for i := 0; i < len(es); i++ {
		err := m.updateEventWithConfigUpdateLock(es[i])
		if err != nil {
			if err != ErrIgnoreChange {
				openlog.Error(fmt.Sprintf("%dth event %+v got error:%v", i, *es[i], err))
			}
			continue
		}
		validEvents = append(validEvents, es[i])
	}

	if len(validEvents) == 0 {
		openlog.Info("all events are invalid")
		return nil
	}

	return m.dispatcher.DispatchModuleEvent(validEvents)
}

func (m *Manager) updateEventWithConfigUpdateLock(e *event.Event) error {
	// refresh all configuration one by one
	if e == nil || e.EventSource == "" || e.Key == "" {
		return errors.New("nil or invalid event supplied")
	}
	if e.HasUpdated {
		openlog.Info(fmt.Sprintf("config update event %+v has been updated", *e))
		return nil
	}
	openlog.Info(fmt.Sprintf("config update event %+v received", e))
	m.configUpdateMux.Lock()
	defer m.configUpdateMux.Unlock()
	src, val := m.findNextBestSource(e.Key, "")
	if src == nil {
		_, ok := m.ConfigurationMap.Load(e.Key)
		if !ok {
			openlog.Info(fmt.Sprintf("the key %s is not existed, ignore", e.Key))
			return ErrIgnoreChange
		}
		m.deleteCacheWithoutConfigUpdateLock(e.Key)
		e.EventType = event.Delete
	} else {
		_, ok := m.ConfigurationMap.Load(e.Key)
		if !ok {
			e.EventType = event.Create
		} else {
			e.EventType = event.Update
		}
		oldVal, ok2 := m.ConfigValueCache.Load(e.Key)
		if ok2 != ok {
			openlog.Error(fmt.Sprintf("unexpected err. ConfigValueCache & ConfigurationMap not consistancy @ key %s", e.Key))
		}
		// we need to update cache if oldSrc != src || oldVal != val
		m.updateCacheWithoutConfigUpdateLock(src.GetSourceName(), e.Key, val)
		if oldVal == val {
			openlog.Info(fmt.Sprintf("the key %s value %s is up-to-date, ignore value change", e.Key, val))
			return ErrIgnoreChange
		}
	}
	e.HasUpdated = true
	return nil
}

func (m *Manager) updateCacheWithoutConfigUpdateLock(source, key string, val interface{}) {
	m.ConfigurationMap.Store(key, source)
	m.ConfigValueCache.Store(key, val)
}

func (m *Manager) deleteCacheWithoutConfigUpdateLock(key string) {
	m.ConfigurationMap.Delete(key)
	m.ConfigValueCache.Delete(key)
}

// OnEvent Triggers actions when an event is generated
func (m *Manager) OnEvent(event *event.Event) {
	err := m.updateEventWithConfigUpdateLock(event)
	if err != nil {
		if err != ErrIgnoreChange {
			openlog.Error("failed in updating event with error: " + err.Error())
		}
		return
	}

	m.dispatcher.DispatchEvent(event)
}

// OnModuleEvent Triggers actions when events are generated
func (m *Manager) OnModuleEvent(event []*event.Event) {
	if err := m.updateModuleEvent(event); err != nil {
		openlog.Error("failed in updating events with error: " + err.Error())
	}
}

func (m *Manager) findNextBestSource(key string, sourceName string) (rSource ConfigSource, value interface{}) {
	m.sourceMapMux.RLock()
	for _, source := range m.Sources {
		if source.GetSourceName() == sourceName {
			continue
		}
		tmpValue, err := source.GetConfigurationByKey(key)
		if err != nil || tmpValue == nil {
			continue
		}
		if rSource == nil {
			rSource, value = source, tmpValue
			continue
		}
		if source.GetPriority() < rSource.GetPriority() { // less value has high priority
			rSource, value = source, tmpValue
		}
	}
	m.sourceMapMux.RUnlock()

	return rSource, value
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
