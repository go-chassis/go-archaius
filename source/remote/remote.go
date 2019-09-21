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

//Package remote created on 2017/6/22.
package remote

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/go-archaius/source"
	"github.com/go-chassis/go-chassis-config"
	"github.com/go-mesh/openlogging"
)

// const
const (
	//Name variable of type string
	Name                       = "ConfigCenterSource"
	configCenterSourcePriority = 0
	ModeInterval               = 1
)

var (
	//ConfigPath is a variable of type string
	ConfigPath = ""
	//ConfigRefreshPath is a variable of type string
	ConfigRefreshPath = ""
)

//Source handles configs from config center
type Source struct {
	cc config.Client

	connsLock sync.Mutex

	dimensions []map[string]string

	sync.RWMutex
	Configurations map[string]interface{}

	dimensionsInfoConfiguration  map[string]map[string]interface{}
	dimensionsInfoConfigurations []map[string]map[string]interface{}

	RefreshMode     int
	RefreshInterval time.Duration
	priority        int

	eh source.EventHandler
}

//NewConfigCenterSource initializes all components of configuration center
func NewConfigCenterSource(cc config.Client, refreshMode, refreshInterval int) source.ConfigSource {
	s := new(Source)
	s.dimensions = []map[string]string{cc.Options().Labels}
	s.priority = configCenterSourcePriority
	s.cc = cc
	s.RefreshMode = refreshMode
	s.RefreshInterval = time.Second * time.Duration(refreshInterval)
	return s
}

//GetConfigurations pull config from remote and start refresh configs interval
// write a new map and return, internal map can not be operated outside struct
func (rs *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	err := rs.refreshConfigurations()
	if err != nil {
		return nil, err
	}
	if rs.RefreshMode == ModeInterval {
		go rs.refreshConfigurationsPeriodically()
	}

	rs.Lock()
	for key, value := range rs.Configurations {
		configMap[key] = value
	}
	rs.Unlock()

	return configMap, nil
}

func (rs *Source) refreshConfigurationsPeriodically() {
	ticker := time.Tick(rs.RefreshInterval)
	for range ticker {
		err := rs.refreshConfigurations()
		if err == nil {
			openlogging.Error("can not pull configs: " + err.Error())
		}
	}
}

func (rs *Source) refreshConfigurations() error {
	var (
		config map[string]interface{}
		err    error
		events []*event.Event
	)

	config, err = rs.cc.PullConfigs(rs.dimensions...)
	if err != nil {
		openlogging.GetLogger().Warnf("Failed to pull configurations from config center server", err) //Warn
		return err
	}
	openlogging.Debug("pull configs", openlogging.WithTags(openlogging.Tags{
		"config": config,
	}))
	rs.Lock()
	rs.Configurations = config
	rs.Unlock()
	//Populate the events based on the changed value between current config and newly received Config
	events, err = rs.populateEvents(config)
	if err != nil {
		openlogging.GetLogger().Warnf("error in generating event", err)
		return err
	}

	//Generate OnEvent Callback based on the events created
	if rs.eh != nil {
		openlogging.GetLogger().Debugf("event on receive %+v", events)
		for _, event := range events {
			rs.eh.OnEvent(event)
		}
	}

	return nil
}

//GetConfigurationByKey gets required configuration for a particular key
func (rs *Source) GetConfigurationByKey(key string) (interface{}, error) {
	rs.Lock()
	configSrcVal, ok := rs.Configurations[key]
	rs.Unlock()
	if ok {
		return configSrcVal, nil
	}

	return nil, errors.New("key not exist")
}

//AddDimensionInfo adds dimension info for a configuration
func (rs *Source) AddDimensionInfo(labels map[string]string) error {
	// TODO check duplication labels
	rs.dimensions = append(rs.dimensions, labels)
	return nil
}

//GetSourceName returns name of the configuration
func (*Source) GetSourceName() string {
	return Name
}

//GetPriority returns priority of a configuration
func (rs *Source) GetPriority() int {
	return rs.priority
}

//SetPriority custom priority
func (rs *Source) SetPriority(priority int) {
	rs.priority = priority
}

//Watch dynamically handles a configuration
func (rs *Source) Watch(callback source.EventHandler) error {
	rs.eh = callback
	if rs.RefreshMode == 0 {
		// Pull All the configuration for the first time.
		rs.refreshConfigurations()
		//Start watch and receive change events.
		rs.cc.Watch(
			func(kv map[string]interface{}) {
				events, err := rs.populateEvents(kv)
				if err != nil {
					openlogging.GetLogger().Error("error in generating event:" + err.Error())
					return
				}

				openlogging.GetLogger().Debugf("event On Receive", events)
				for _, event := range events {
					callback.OnEvent(event)
				}

				return
			},
			func(err error) {
				openlogging.Error(err.Error())
			}, nil,
		)
	}

	return nil
}

//Cleanup cleans the particular configuration up
func (rs *Source) Cleanup() error {
	rs.connsLock.Lock()
	defer rs.connsLock.Unlock()

	rs.Configurations = nil

	return nil
}

func (rs *Source) populateEvents(updatedConfig map[string]interface{}) ([]*event.Event, error) {
	events := make([]*event.Event, 0)
	newConfig := make(map[string]interface{})
	rs.Lock()
	defer rs.Unlock()

	currentConfig := rs.Configurations

	// generate create and update event
	for key, value := range updatedConfig {
		newConfig[key] = value
		currentValue, ok := currentConfig[key]
		if !ok { // if new configuration introduced
			events = append(events, rs.constructEvent(event.Create, key, value))
		} else if !reflect.DeepEqual(currentValue, value) {
			events = append(events, rs.constructEvent(event.Update, key, value))
		}
	}

	// generate delete event
	for key, value := range currentConfig {
		_, ok := newConfig[key]
		if !ok { // when old config not present in new config
			events = append(events, rs.constructEvent(event.Delete, key, value))
		}
	}

	// update with latest config
	rs.Configurations = newConfig

	return events, nil
}

func (rs *Source) setKeyValueByDI(updatedConfig map[string]map[string]interface{}, dimensionInfo string) ([]*event.Event, error) {
	events := make([]*event.Event, 0)
	newConfigForDI := make(map[string]map[string]interface{})
	rs.Lock()
	defer rs.Unlock()

	currentConfig := rs.dimensionsInfoConfiguration

	// generate create and update event
	for key, value := range updatedConfig {
		if key == dimensionInfo {
			newConfigForDI[key] = value
			for k, v := range value {
				if len(currentConfig) == 0 {
					events = append(events, rs.constructEvent(event.Create, k, v))
				}
				for diKey, val := range currentConfig {
					if diKey == dimensionInfo {
						currentValue, ok := val[k]
						if !ok { // if new configuration introduced
							events = append(events, rs.constructEvent(event.Create, k, v))
						} else if currentValue != v {
							events = append(events, rs.constructEvent(event.Update, k, v))
						}
					}
				}
			}
		}
	}

	// generate delete event
	for key, value := range currentConfig {
		if key == dimensionInfo {
			for k, v := range value {
				for _, val := range newConfigForDI {
					_, ok := val[k]
					if !ok {
						events = append(events, rs.constructEvent(event.Delete, k, v))
					}
				}
			}
		}
	}

	// update with latest config
	rs.dimensionsInfoConfiguration = newConfigForDI

	return events, nil
}

func (rs *Source) constructEvent(eventType string, key string, value interface{}) *event.Event {
	newEvent := new(event.Event)
	newEvent.EventSource = Name
	newEvent.EventType = eventType
	newEvent.Key = key
	newEvent.Value = value

	return newEvent
}

//Set no use
func (rs *Source) Set(key string, value interface{}) error {
	return nil
}

//Delete no use
func (rs *Source) Delete(key string) error {
	return nil
}
