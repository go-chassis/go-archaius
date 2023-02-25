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
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/arielsrv/go-archaius/source/remote"
	"github.com/arielsrv/go-archaius/source/util/queue"
	client "github.com/go-chassis/kie-client"
	"github.com/go-chassis/openlog"
)

// DimensionName is the name identifying various dimension of configurations.
// One dimension corresponds to one label combination.
type DimensionName string

// Priority of dimension in dimensionPrecedence array must from low to high
var dimensionPrecedence = []DimensionName{
	DimensionApp,
	DimensionService,
}

// const
const (
	defaultWaitTime                = 30
	DimensionApp     DimensionName = "app"
	DimensionService DimensionName = "service"
)

// Kie is Implementation
type Kie struct {
	c    *client.Client
	opts remote.Options
	//wait time for watching, unit is second
	watchTimeOut int

	currentRevision int

	dimensions map[DimensionName]*Dimension
}

// Dimension contains a label combination and the configuration corresponding to this label combination
type Dimension struct {
	sync.RWMutex
	labels map[string]string
	config map[string]interface{}
}

// NewKie is a return a new kie client
func NewKie(options remote.Options) (*Kie, error) {
	if options.ServerURI == "" {
		return nil, remote.ErrInvalidEP
	}
	kies := strings.Split(options.ServerURI, ",")
	ks := make([]string, 0)
	for _, value := range kies {
		value = strings.Replace(value, " ", "", -1)
		ks = append(ks, value)
	}

	c, err := client.NewClient(client.Config{
		Endpoint:   ks[0],
		VerifyPeer: options.VerifyPeer,
	})
	if err != nil {
		return nil, err
	}
	dimensions, err := initDimensions(options.Labels)
	if err != nil {
		return nil, err
	}

	kie := &Kie{
		c:               c,
		opts:            options,
		watchTimeOut:    options.WatchTimeOut,
		currentRevision: -1,
		dimensions:      dimensions,
	}
	openlog.Info("new kie client", openlog.WithTags(
		openlog.Tags{
			"verifyPeer": options.VerifyPeer,
			"ep":         ks,
		}))
	return kie, nil
}

func initDimensions(optionsLabels map[string]string) (map[DimensionName]*Dimension, error) {
	dimensions := make(map[DimensionName]*Dimension)
	for _, dimension := range dimensionPrecedence {
		labels, err := GenerateLabels(dimension, optionsLabels)
		if err != nil {
			return nil, err
		}
		dimensions[dimension] = &Dimension{
			labels: labels,
			config: make(map[string]interface{}),
		}
	}
	return dimensions, nil
}

// PullConfigs is the implementation of Kie to pull all the configurations from Config-Server
func (k *Kie) PullConfigs(labels ...map[string]string) (map[string]interface{}, error) {
	var revisionLock sync.Mutex
	var validRevisions []int
	getKVDimensionally := func(i int, errCh chan error) {
		kv, responseRevision, err := k.c.List(context.Background(),
			client.WithGetProject(k.opts.ProjectID),
			client.WithLabels(k.getDimensionLabels(dimensionPrecedence[i])),
			client.WithExact(),
			client.WithRevision(k.currentRevision))
		if responseRevision >= 0 {
			revisionLock.Lock()
			validRevisions = append(validRevisions, responseRevision)
			revisionLock.Unlock()
		}
		if err != nil && err != client.ErrKeyNotExist {
			//If the error is the no changes error, return immediately,
			//otherwise append the error
			if err != client.ErrNoChanges {
				errCh <- err
			}
			return
		}
		//If found an updated kv or no kv was found, then update configs cache
		k.setDimensionConfigs(kv, dimensionPrecedence[i])
	}
	err := queue.Concurrent(len(dimensionPrecedence), len(dimensionPrecedence), getKVDimensionally)
	if err != nil {
		return nil, err
	}
	//Find the minimum valid revision from responses. The next pull request will use it.
	if len(validRevisions) > 0 {
		currentRevision := validRevisions[0]
		for _, revision := range validRevisions {
			if revision < currentRevision {
				currentRevision = revision
			}
		}
		k.currentRevision = currentRevision
	}
	return k.mergeConfig(), nil
}

// Watch watch the configuration changes and update in real time
func (k *Kie) Watch(f func(map[string]interface{}), errHandler func(err error), labels map[string]string) error {
	for _, dimension := range dimensionPrecedence {
		go k.watchKVDimensionally(f, errHandler, dimension)
	}
	return nil
}

func (k *Kie) watchKVDimensionally(f func(map[string]interface{}), errHandler func(err error), dimension DimensionName) {
	openlog.Info("start watching configurations of dimension " + string(dimension))
	defer openlog.Info("stop watching configurations of dimension " + string(dimension))
	if k.watchTimeOut == 0 {
		k.watchTimeOut = defaultWaitTime
	}
	wait := fmt.Sprintf("%ds", k.watchTimeOut)
	revision := -1
	for {
		kv, responseRevision, err := k.c.List(context.Background(),
			client.WithGetProject(k.opts.ProjectID),
			client.WithLabels(k.getDimensionLabels(dimension)),
			client.WithExact(),
			client.WithRevision(revision),
			client.WithWait(wait))
		if responseRevision >= 0 {
			revision = responseRevision
		}
		if err != nil && err != client.ErrKeyNotExist {
			//If the error is the no changes error, execute the next watch immediately,
			//otherwise print the error and wait for some time.
			if err != client.ErrNoChanges {
				errHandler(err)
				time.Sleep(time.Second * time.Duration(k.watchTimeOut))
			}
			continue
		}
		//If found an updated kv or no kv was found, then update configs cache
		if updated := k.setDimensionConfigs(kv, dimension); updated {
			f(k.mergeConfig())
		}
	}
}

func (k *Kie) setDimensionConfigs(kvs *client.KVResponse, dimension DimensionName) bool {
	if k.dimensions[dimension] == nil {
		return false
	}
	configs := make(map[string]interface{})
	if kvs == nil {
		return false
	}
	for _, kv := range kvs.Data {
		if kv == nil {
			continue
		}
		k := kv.Key
		if k != "" && kv.Status == "enabled" {
			configs[k] = kv.Value
		}
	}

	k.dimensions[dimension].Lock()
	defer k.dimensions[dimension].Unlock()
	if len(k.dimensions[dimension].config) == 0 && len(configs) == 0 {
		//kv neither found in remote nor in local, no need to update
		return false
	}
	k.dimensions[dimension].config = configs
	return true
}

func (k *Kie) mergeConfig() map[string]interface{} {
	configs := make(map[string]interface{})
	if k.dimensions != nil {
		for _, dimension := range dimensionPrecedence {
			if k.dimensions[dimension] != nil {
				k.dimensions[dimension].RLock()
				for k, v := range k.dimensions[dimension].config {
					configs[k] = v
				}
				k.dimensions[dimension].RUnlock()
			}
		}
	}
	return configs
}

func (k *Kie) getDimensionLabels(dimension DimensionName) map[string]string {
	if k.dimensions[dimension] == nil {
		return nil
	}
	return k.dimensions[dimension].labels
}

// Options return options
func (k *Kie) Options() remote.Options {
	return k.opts
}
