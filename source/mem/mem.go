package mem

import (
	"errors"
	"sync"

	"github.com/arielsrv/go-archaius/event"

	"github.com/arielsrv/go-archaius/source"
	"github.com/go-chassis/openlog"
)

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

// const
const (
	Name                         = "MemorySource"
	memoryVariableSourcePriority = 1
)

var ErrSourceNotReady = errors.New("source is not ready")

// Source is a struct
type Source struct {
	Configs sync.Map

	callback source.EventHandler
	waitOnce sync.Once
	Ready    chan bool
	priority int
}

// NewMemoryConfigurationSource initializes all necessary components for memory configuration
func NewMemoryConfigurationSource() source.ConfigSource {
	memoryConfigSource := new(Source)
	memoryConfigSource.priority = memoryVariableSourcePriority
	memoryConfigSource.Configs = sync.Map{}
	memoryConfigSource.Ready = make(chan bool)
	return memoryConfigSource
}

// GetConfigurations gets all memory configurations
func (ms *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	ms.Configs.Range(func(key, value interface{}) bool {
		configMap[key.(string)] = value
		return true
	})

	return configMap, nil
}

// GetConfigurationByKey gets required memory configuration for a particular key
func (ms *Source) GetConfigurationByKey(key string) (interface{}, error) {
	value, ok := ms.Configs.Load(key)
	if !ok {
		return nil, source.ErrKeyNotExist
	}

	return value, nil
}

// GetPriority returns priority of the memory configuration
func (ms *Source) GetPriority() int {
	return ms.priority
}

// SetPriority custom priority
func (ms *Source) SetPriority(priority int) {
	ms.priority = priority
}

// GetSourceName returns name of memory configuration
func (*Source) GetSourceName() string {
	return Name
}

// Watch dynamically handles a memory configuration
func (ms *Source) Watch(callback source.EventHandler) error {
	ms.callback = callback
	openlog.Info("mem source callback prepared")
	ms.Ready <- true
	return nil
}

// Cleanup cleans a particular memory configuration up
func (ms *Source) Cleanup() error {
	ms.Configs = sync.Map{}
	return nil
}

// AddDimensionInfo  is none function
func (ms *Source) AddDimensionInfo(labels map[string]string) error {
	return nil
}

// Set set mem config
func (ms *Source) Set(key string, value interface{}) error {
	ms.waitOnce.Do(func() {
		<-ms.Ready
	})

	e := new(event.Event)
	e.EventSource = ms.GetSourceName()
	e.Key = key
	e.Value = value

	if _, ok := ms.Configs.Load(key); !ok {
		e.EventType = event.Create
	} else {
		e.EventType = event.Update
	}

	ms.Configs.Store(key, value)

	if ms.callback != nil {
		ms.callback.OnEvent(e)
		ms.callback.OnModuleEvent([]*event.Event{e})
	}

	return nil
}

// Delete remvove mem config
func (ms *Source) Delete(key string) error {
	ms.waitOnce.Do(func() {
		<-ms.Ready
	})

	e := new(event.Event)
	e.EventSource = ms.GetSourceName()
	e.Key = key

	if v, ok := ms.Configs.Load(key); ok {
		e.EventType = event.Delete
		e.Value = v
	} else {
		return nil
	}

	if ms.callback != nil {
		ms.callback.OnEvent(e)
		ms.callback.OnModuleEvent([]*event.Event{e})
	}

	return nil
}
