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
	"strings"

	"github.com/arielsrv/go-archaius/pkg/configcenter"
	"github.com/arielsrv/go-archaius/source/remote"
	"github.com/go-chassis/openlog"
	"github.com/gorilla/websocket"
)

const (
	//HeaderContentType is a variable of type string
	HeaderContentType = "Content-Type"
	//HeaderUserAgent is a variable of type string
	HeaderUserAgent = "User-Agent"
)

// ConfigCenter is Implementation
type ConfigCenter struct {
	c        *configcenter.Client
	opts     remote.Options
	wsDialer *websocket.Dialer
}

// NewConfigCenter is a function
func NewConfigCenter(options remote.Options) (*ConfigCenter, error) {
	if options.ServerURI == "" {
		return nil, remote.ErrInvalidEP
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
	openlog.Info("new config center client", openlog.WithTags(
		openlog.Tags{
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

// Watch use ws
func (c *ConfigCenter) Watch(f func(map[string]interface{}), errHandler func(err error), labels map[string]string) error {
	return c.c.Watch(f, errHandler)
}

// Options return options
func (c *ConfigCenter) Options() remote.Options {
	return c.opts
}
