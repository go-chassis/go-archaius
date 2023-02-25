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

// Package env created on 2017/6/22.
package env

import (
	"os"
	"strings"
	"sync"

	"github.com/arielsrv/go-archaius/source"

	"github.com/go-chassis/openlog"
)

const (
	envSourceConst            = "EnvironmentSource"
	envVariableSourcePriority = 3
)

// Source is a struct
type Source struct {
	Configs  sync.Map
	priority int
}

// NewEnvConfigurationSource configures a new environment configuration
func NewEnvConfigurationSource() source.ConfigSource {
	openlog.Info("enable env source")
	envConfigSource := new(Source)
	envConfigSource.priority = envVariableSourcePriority
	envConfigSource.pullConfigurations()
	return envConfigSource
}

func (es *Source) pullConfigurations() {
	es.Configs = sync.Map{}
	for _, value := range os.Environ() {
		rs := []rune(value)
		in := strings.Index(value, "=")
		key := string(rs[0:in])
		value := string(rs[in+1:])
		envKey := strings.Replace(key, "_", ".", -1)
		es.Configs.Store(key, value)
		es.Configs.Store(envKey, value)

	}
}

// GetConfigurations gets all configuration
func (es *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	es.Configs.Range(func(k, v interface{}) bool {
		configMap[k.(string)] = v
		return true
	})

	return configMap, nil
}

// GetConfigurationByKey gets required configuration for a particular key
func (es *Source) GetConfigurationByKey(key string) (interface{}, error) {
	value, ok := es.Configs.Load(key)
	if !ok {
		return nil, source.ErrKeyNotExist
	}

	return value, nil
}

// GetPriority returns priority of environment configuration
func (es *Source) GetPriority() int {
	return es.priority
}

// SetPriority custom priority
func (es *Source) SetPriority(priority int) {
	es.priority = priority
}

// GetSourceName returns the name of environment source
func (*Source) GetSourceName() string {
	return envSourceConst
}

// Watch dynamically handles a environment configuration
func (*Source) Watch(callback source.EventHandler) error {
	//TODO env change
	return nil
}

// Cleanup cleans a particular environment configuration up
func (es *Source) Cleanup() error {
	es.Configs = sync.Map{}
	return nil
}

// AddDimensionInfo no use
func (es *Source) AddDimensionInfo(labels map[string]string) error {
	return nil
}

// Set no use
func (es *Source) Set(key string, value interface{}) error {
	return nil
}

// Delete no use
func (es *Source) Delete(key string) error {
	return nil
}
