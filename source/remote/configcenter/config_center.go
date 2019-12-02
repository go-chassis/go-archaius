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

// Package configcenter created on 2017/6/20.
package configcenter

import (
	"errors"
	"github.com/go-chassis/go-archaius/pkg/configcenter"
	"github.com/go-chassis/go-archaius/source/remote"
	"strings"

	"github.com/go-mesh/openlogging"
	"github.com/gorilla/websocket"
)

const (
	//HeaderContentType is a variable of type string
	HeaderContentType = "Content-Type"
	//HeaderUserAgent is a variable of type string
	HeaderUserAgent = "User-Agent"
	// Name of the Plugin
	Name = "config_center"
)

//errors
var (
	ErrInvalidEP = errors.New("invalid endpoint")
)

//ConfigCenter is Implementation
type ConfigCenter struct {
	c        *configcenter.Client
	opts     remote.Options
	wsDialer *websocket.Dialer
}

//NewConfigCenter is a function
func NewConfigCenter(options remote.Options) (remote.Client, error) {
	if options.ServerURI == "" {
		return nil, ErrInvalidEP
	}
	configCenters := strings.Split(options.ServerURI, ",")
	cCenters := make([]string, 0)
	for _, value := range configCenters {
		value = strings.Replace(value, " ", "", -1)
		cCenters = append(cCenters, value)
	}
	d, err := GenerateDimension(options.Labels[remote.LabelService], options.Labels[remote.LabelVersion], options.Labels[remote.LabelApp])
	if err != nil {
		return nil, err
	}

	c, err := configcenter.New(configcenter.Options{
		ConfigServerAddresses: cCenters,
		DefaultDimension:      d,
		TLSConfig:             options.TLSConfig,
		TenantName:            options.TenantName,
		EnableSSL:             options.EnableSSL,
		RefreshPort:           options.RefreshPort,
	})
	if err != nil {
		return nil, err
	}

	cc := &ConfigCenter{
		c:    c,
		opts: options,
	}
	openlogging.Info("new config center client", openlogging.WithTags(
		openlogging.Tags{
			"dimension": d,
			"ws_port":   options.RefreshPort,
			"ssl":       options.EnableSSL,
			"ep":        cCenters,
		}))
	return cc, nil
}

// PullConfigs is the implementation of ConfigCenter to pull all the configurations from Config-Server
func (c *ConfigCenter) PullConfigs(labels ...map[string]string) (map[string]interface{}, error) {
	d := ""
	var err error
	d, err = GenerateDimension(c.opts.Labels[remote.LabelService], c.opts.Labels[remote.LabelVersion], c.opts.Labels[remote.LabelApp])
	if err != nil {
		return nil, err
	}
	configurations, error := c.c.Flatten(d)
	if error != nil {
		return nil, error
	}
	return configurations, nil
}

// PullConfig is the implementation of ConfigCenter to pull specific configurations from Config-Server
func (c *ConfigCenter) PullConfig(key, contentType string, labels map[string]string) (interface{}, error) {
	if len(labels) == 0 {
		labels = c.opts.Labels
	}
	d, err := GenerateDimension(c.opts.Labels[remote.LabelService], c.opts.Labels[remote.LabelVersion], c.opts.Labels[remote.LabelApp])
	if err != nil {
		return nil, err
	}
	// TODO use the contentType to return the configurations
	configurations, error := c.c.Flatten(d)
	if error != nil {
		return nil, error
	}
	configurationsValue, ok := configurations[key]
	if !ok {
		openlogging.GetLogger().Error("Error in fetching the configurations for particular value,No Key found : " + key)
	}

	return configurationsValue, nil
}

// PushConfigs push configs to ConfigSource cc , success will return { "Result": "Success" }
func (c *ConfigCenter) PushConfigs(items map[string]interface{}, labels map[string]string) (map[string]interface{}, error) {
	if len(items) == 0 {
		em := "data is empty , which data need to send cc"
		openlogging.GetLogger().Error(em)
		return nil, errors.New(em)
	}
	if len(labels) == 0 {
		labels = c.opts.Labels
	}
	d, err := GenerateDimension(c.opts.Labels[remote.LabelService], c.opts.Labels[remote.LabelVersion], c.opts.Labels[remote.LabelApp])
	if err != nil {
		return nil, err
	}
	configAPI := &configcenter.CreateConfigAPI{
		DimensionInfo: d,
		Items:         items,
	}

	return c.c.AddConfig(configAPI)
}

// DeleteConfigsByKeys delete keys
func (c *ConfigCenter) DeleteConfigsByKeys(keys []string, labels map[string]string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		em := "not key need to delete for cc, please check keys"
		openlogging.GetLogger().Error(em)
		return nil, errors.New(em)
	}
	if len(labels) == 0 {
		labels = c.opts.Labels
	}
	d, err := GenerateDimension(c.opts.Labels[remote.LabelService], c.opts.Labels[remote.LabelVersion], c.opts.Labels[remote.LabelApp])
	if err != nil {
		return nil, err
	}
	configAPI := &configcenter.DeleteConfigAPI{
		DimensionInfo: d,
		Keys:          keys,
	}

	return c.c.DeleteConfig(configAPI)
}

//Watch use ws
func (c *ConfigCenter) Watch(f func(map[string]interface{}), errHandler func(err error), labels map[string]string) error {
	return c.c.Watch(f, errHandler)
}
func init() {
	remote.InstallConfigClientPlugin(Name, NewConfigCenter)
}

//Options return options
func (c *ConfigCenter) Options() remote.Options {
	return c.opts
}
