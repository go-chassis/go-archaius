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

package kie

import (
	"errors"
	"sync"
	"time"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/go-archaius/source"
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/go-mesh/openlogging"
)

// const
const (
	//Name is the source name of kie
	Name              = "KieSource"
	kieSourcePriority = 0
)

//Source handles configs from ServiceComb-Kie
type Source struct {
	k *Kie

	dimensions []map[string]string

	sync.RWMutex
	currentConfig map[string]interface{}

	RefreshMode     int
	RefreshInterval time.Duration
	priority        int

	eh source.EventHandler
}

//NewKieSource initializes all components of ServiceComb-Kie
func NewKieSource(ci *archaius.RemoteInfo) (source.ConfigSource, error) {
	opts := remote.Options{
		ServerURI:     ci.URL,
		TenantName:    ci.TenantName,
		EnableSSL:     ci.EnableSSL,
		TLSConfig:     ci.TLSConfig,
		RefreshPort:   ci.RefreshPort,
		AutoDiscovery: ci.AutoDiscovery,
		Labels:        ci.DefaultDimension,
		WatchTimeOut:  ci.RefreshInterval,
	}
	k, err := NewKie(opts)
	if err != nil {
		openlogging.Error(err.Error())
		return nil, err
	}
	ks := new(Source)
	ks.dimensions = []map[string]string{k.Options().Labels}
	ks.priority = kieSourcePriority
	ks.k = k
	ks.RefreshMode = ci.RefreshMode
	if ci.RefreshInterval == 0 {
		ks.RefreshInterval = remote.DefaultInterval
	} else {
		ks.RefreshInterval = time.Second * time.Duration(ci.RefreshInterval)
	}
	return ks, nil
}

//GetConfigurations pull config from remote and start refresh configs interval
// write a new map and return, internal map can not be operated outside struct
func (ks *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	err := ks.refreshConfigurations()
	if err != nil {
		return nil, err
	}
	if ks.RefreshMode == remote.ModeInterval {
		go ks.refreshConfigurationsPeriodically()
	}

	ks.RLock()
	for key, value := range ks.currentConfig {
		configMap[key] = value
	}
	ks.RUnlock()

	return configMap, nil
}

func (ks *Source) refreshConfigurationsPeriodically() {
	ticker := time.Tick(ks.RefreshInterval)
	openlogging.Info("start refreshing configurations")
	for range ticker {
		err := ks.refreshConfigurations()
		if err != nil {
			openlogging.Error("can not pull configs: " + err.Error())
		}
	}
	openlogging.Info("stop refreshing configurations")
}

func (ks *Source) refreshConfigurations() error {
	config, err := ks.k.PullConfigs(ks.dimensions...)
	if err != nil {
		openlogging.GetLogger().Warnf("Failed to pull configurations from kie server", err) //Warn
		return err
	}
	openlogging.Debug("pull configs from kie", openlogging.WithTags(openlogging.Tags{
		"config": config,
	}))
	return ks.updateConfigAndFireEvent(config)
}

func (ks *Source) updateConfigAndFireEvent(config map[string]interface{}) error {
	ks.Lock()
	defer ks.Unlock()
	//Populate the events based on the changed value between current config and newly received Config
	events, err := event.PopulateEvents(Name, ks.currentConfig, config)
	if err != nil {
		openlogging.GetLogger().Warnf("generating event error", err)
		return err
	}
	ks.currentConfig = config
	//Generate OnEvent Callback based on the events created
	if ks.eh != nil {
		openlogging.GetLogger().Debugf("received event %s", events)
		for _, e := range events {
			ks.eh.OnEvent(e)
		}
	}
	return nil
}

//GetConfigurationByKey gets required configuration for a particular key
func (ks *Source) GetConfigurationByKey(key string) (interface{}, error) {
	if ks.currentConfig == nil {
		return nil, errors.New("currentConfig is nil")
	}
	ks.RLock()
	configSrcVal, ok := ks.currentConfig[key]
	ks.RUnlock()
	if ok {
		return configSrcVal, nil
	}

	return nil, source.ErrKeyNotExist
}

//AddDimensionInfo adds dimension info for a configuration
func (ks *Source) AddDimensionInfo(labels map[string]string) error {
	// TODO check duplication labels
	ks.dimensions = append(ks.dimensions, labels)
	return nil
}

//GetSourceName returns name of the configuration
func (*Source) GetSourceName() string {
	return Name
}

//GetPriority returns priority of a configuration
func (ks *Source) GetPriority() int {
	return ks.priority
}

//SetPriority custom priority
func (ks *Source) SetPriority(priority int) {
	ks.priority = priority
}

//Watch dynamically handles a configuration
func (ks *Source) Watch(callback source.EventHandler) error {
	ks.eh = callback
	if ks.RefreshMode != remote.ModeWatch {
		return nil
	}
	//Start watch and receive change events.
	openlogging.Info("start watching configurations")
	err := ks.k.Watch(
		func(kv map[string]interface{}) {
			openlogging.Debug("watch configs", openlogging.WithTags(openlogging.Tags{
				"config": kv,
			}))
			err := ks.updateConfigAndFireEvent(kv)
			if err != nil {
				openlogging.GetLogger().Error("error in updating configurations:" + err.Error())
			}
		},
		func(err error) {
			openlogging.Error(err.Error())
		}, nil,
	)
	openlogging.Info("stop watching configurations")
	if err != nil {
		openlogging.Error("watch kie source failed: " + err.Error())
	}
	return err
}

//Cleanup cleans the particular configuration up
func (ks *Source) Cleanup() error {
	ks.Lock()
	defer ks.Unlock()

	ks.currentConfig = nil

	return nil
}

//Set no use
func (ks *Source) Set(key string, value interface{}) error {
	return nil
}

//Delete no use
func (ks *Source) Delete(key string) error {
	return nil
}
func init() {
	archaius.InstallRemoteSource(archaius.KieSource, NewKieSource)
}
