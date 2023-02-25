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

package configcenter

import (
	"fmt"
	"sync"
	"time"

	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/event"
	"github.com/arielsrv/go-archaius/source"
	"github.com/arielsrv/go-archaius/source/remote"
	"github.com/go-chassis/openlog"
)

// const
const (
	//ConfigCenterSourceName variable of type string
	ConfigCenterSourceName     = "ConfigCenterSource"
	configCenterSourcePriority = 0
)

// Source handles configs from config center
type Source struct {
	c *ConfigCenter

	connsLock sync.Mutex

	dimensions []map[string]string

	sync.RWMutex
	currentConfig map[string]interface{}

	dimensionsInfoConfiguration  map[string]map[string]interface{}
	dimensionsInfoConfigurations []map[string]map[string]interface{}

	RefreshMode     int
	RefreshInterval time.Duration
	priority        int

	eh source.EventHandler
}

// NewConfigCenterSource initializes all components of configuration center
func NewConfigCenterSource(ci *archaius.RemoteInfo) (source.ConfigSource, error) {
	opts := remote.Options{
		ServerURI:     ci.URL,
		TenantName:    ci.TenantName,
		EnableSSL:     ci.EnableSSL,
		TLSConfig:     ci.TLSConfig,
		RefreshPort:   ci.RefreshPort,
		AutoDiscovery: ci.AutoDiscovery,
		Labels:        ci.DefaultDimension,
	}
	cc, err := NewConfigCenter(opts)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	s := new(Source)
	s.dimensions = []map[string]string{cc.Options().Labels}
	s.priority = configCenterSourcePriority
	s.c = cc
	s.RefreshMode = ci.RefreshMode
	s.RefreshInterval = time.Second * time.Duration(ci.RefreshInterval)
	return s, nil
}

// GetConfigurations pull config from remote and start refresh configs interval
// write a new map and return, internal map can not be operated outside struct
func (rs *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	err := rs.refreshConfigurations()
	if err != nil {
		return nil, err
	}
	if rs.RefreshMode == remote.ModeInterval {
		go rs.refreshConfigurationsPeriodically()
	}

	rs.Lock()
	for key, value := range rs.currentConfig {
		configMap[key] = value
	}
	rs.Unlock()

	return configMap, nil
}

func (rs *Source) refreshConfigurationsPeriodically() {
	ticker := time.Tick(rs.RefreshInterval)
	for range ticker {
		err := rs.refreshConfigurations()
		if err != nil {
			openlog.Error("can not pull configs: " + err.Error())
		}
	}
}

func (rs *Source) refreshConfigurations() error {
	var (
		config map[string]interface{}
		err    error
		events []*event.Event
	)

	config, err = rs.c.PullConfigs(rs.dimensions...)
	if err != nil {
		openlog.Warn(fmt.Sprintf("failed to pull configurations from config center server %s", err)) //Warn
		return err
	}
	openlog.Debug("pull configs", openlog.WithTags(openlog.Tags{
		"config": config,
	}))
	//Populate the events based on the changed value between current config and newly received Config
	rs.Lock()
	defer rs.Unlock()
	events, err = event.PopulateEvents(ConfigCenterSourceName, rs.currentConfig, config)
	if err != nil {
		openlog.Warn(fmt.Sprintf("error in generating event %s", err))
		return err
	}
	rs.currentConfig = config
	//Generate OnEvent Callback based on the events created
	if rs.eh != nil {
		openlog.Debug(fmt.Sprintf("event on receive %v", events))
		for _, e := range events {
			rs.eh.OnEvent(e)
		}
	}

	return nil
}

// GetConfigurationByKey gets required configuration for a particular key
func (rs *Source) GetConfigurationByKey(key string) (interface{}, error) {
	rs.RLock()
	configSrcVal, ok := rs.currentConfig[key]
	rs.RUnlock()
	if ok {
		return configSrcVal, nil
	}

	return nil, source.ErrKeyNotExist
}

// AddDimensionInfo adds dimension info for a configuration
func (rs *Source) AddDimensionInfo(labels map[string]string) error {
	// TODO check duplication labels
	rs.dimensions = append(rs.dimensions, labels)
	return nil
}

// GetSourceName returns name of the configuration
func (*Source) GetSourceName() string {
	return ConfigCenterSourceName
}

// GetPriority returns priority of a configuration
func (rs *Source) GetPriority() int {
	return rs.priority
}

// SetPriority custom priority
func (rs *Source) SetPriority(priority int) {
	rs.priority = priority
}

// Watch dynamically handles a configuration
func (rs *Source) Watch(callback source.EventHandler) error {
	rs.eh = callback
	if rs.RefreshMode == remote.ModeWatch {
		// Pull All the configuration for the first time.
		rs.refreshConfigurations()
		//Start watch and receive change events.
		err := rs.c.Watch(
			func(kv map[string]interface{}) {
				rs.RLock()
				defer rs.RUnlock()
				events, err := event.PopulateEvents(ConfigCenterSourceName, rs.currentConfig, kv)
				if err != nil {
					openlog.Error("error in generating event:" + err.Error())
					return
				}

				openlog.Debug(fmt.Sprintf("event on receive %v", events))
				for _, e := range events {
					callback.OnEvent(e)
				}

				return
			},
			func(err error) {
				openlog.Error(err.Error())
			}, nil,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// Cleanup cleans the particular configuration up
func (rs *Source) Cleanup() error {
	rs.connsLock.Lock()
	defer rs.connsLock.Unlock()

	rs.currentConfig = nil

	return nil
}

// Set no use
func (rs *Source) Set(key string, value interface{}) error {
	return nil
}

// Delete no use
func (rs *Source) Delete(key string) error {
	return nil
}
func init() {
	archaius.InstallRemoteSource(archaius.ConfigCenterSource, NewConfigCenterSource)
}
