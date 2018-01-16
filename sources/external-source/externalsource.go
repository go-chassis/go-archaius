package externalconfigsource

import (
	"errors"
	"sync"

	"github.com/ServiceComb/go-archaius/core"
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

//Package externalconfigsource created on 2017/6/22.

const (
	externalSourceConst            = "ExternalSource"
	externalVariableSourcePriority = 4
)

//ExternalConfigurationSource is a struct
type ExternalConfigurationSource struct {
	Configurations map[string]interface{}
	callback       core.DynamicConfigCallback
	sync.RWMutex
	CallbackCheck chan bool
	ChanStatus    bool
}

//ExternalSource is a interface
type ExternalSource interface {
	core.ConfigSource
	AddKeyValue(string, interface{}) error
}

var _ core.ConfigSource = &ExternalConfigurationSource{}

var externalConfigSource *ExternalConfigurationSource

//NewExternalConfigurationSource initializes all necessary components for external configuration
func NewExternalConfigurationSource() ExternalSource {
	if externalConfigSource == nil {
		externalConfigSource = new(ExternalConfigurationSource)
		externalConfigSource.Configurations = make(map[string]interface{})
		externalConfigSource.CallbackCheck = make(chan bool)
	}

	return externalConfigSource
}

//GetConfigurations gets all external configurations
func (confSrc *ExternalConfigurationSource) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	confSrc.Lock()
	defer confSrc.Unlock()
	for key, value := range confSrc.Configurations {
		configMap[key] = value
	}

	return configMap, nil
}

//GetConfigurationByKey gets required external configuration for a particular key
func (confSrc *ExternalConfigurationSource) GetConfigurationByKey(key string) (interface{}, error) {
	confSrc.Lock()
	defer confSrc.Unlock()
	value, ok := confSrc.Configurations[key]
	if !ok {
		return nil, errors.New("key does not exist")
	}

	return value, nil
}

//GetPriority returns priority of the external configuration
func (*ExternalConfigurationSource) GetPriority() int {
	return externalVariableSourcePriority
}

//GetSourceName returns name of external configuration
func (*ExternalConfigurationSource) GetSourceName() string {
	return externalSourceConst
}

//DynamicConfigHandler dynamically handles a external configuration
func (confSrc *ExternalConfigurationSource) DynamicConfigHandler(callback core.DynamicConfigCallback) error {
	confSrc.callback = callback
	confSrc.CallbackCheck <- true
	return nil
}

//AddKeyValue creates new configuration for corresponding key and value pair
func (confSrc *ExternalConfigurationSource) AddKeyValue(key string, value interface{}) error {
	if !confSrc.ChanStatus {
		<-confSrc.CallbackCheck
		confSrc.ChanStatus = true
	}

	event := new(core.Event)
	event.EventSource = confSrc.GetSourceName()
	event.Key = key
	event.Value = value

	confSrc.Lock()
	if _, ok := confSrc.Configurations[key]; !ok {
		event.EventType = core.Create
	} else {
		event.EventType = core.Update
	}

	confSrc.Configurations[key] = value
	confSrc.Unlock()

	if confSrc.callback != nil {
		confSrc.callback.OnEvent(event)
	}

	return nil
}

//Cleanup cleans a particular external configuration up
func (confSrc *ExternalConfigurationSource) Cleanup() error {
	confSrc.Configurations = nil

	return nil
}

//GetConfigurationByKeyAndDimensionInfo gets a required external configuration for particular key and dimension info pair
func (*ExternalConfigurationSource) GetConfigurationByKeyAndDimensionInfo(key, di string) (interface{}, error) {
	return nil, nil
}

//AddDimensionInfo adds dimension info for a external configuration
func (*ExternalConfigurationSource) AddDimensionInfo(dimensionInfo string) (map[string]string, error) {
	return nil, nil
}

//GetConfigurationsByDI gets required external configuration for a particular dimension info
func (ExternalConfigurationSource) GetConfigurationsByDI(dimensionInfo string) (map[string]interface{}, error) {
	return nil, nil
}
