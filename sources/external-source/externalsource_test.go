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
package externalconfigsource

import (
	"github.com/ServiceComb/go-archaius/core"
	"testing"
)

type EListener struct {
	Name      string
	EventName string
}

type TestDynamicConfigHandler struct {
	EventName  string
	EventKey   string
	EventValue interface{}
}

func (t *TestDynamicConfigHandler) OnEvent(e *core.Event) {
	t.EventKey = e.Key
	t.EventName = e.EventType
	t.EventValue = e.Value
}

func TestExternalConfigurationSource(t *testing.T) {

	externalsource := NewExternalConfigurationSource()

	t.Log("Test externalsource.go")

	dynHandler := new(TestDynamicConfigHandler)

	go externalsource.DynamicConfigHandler(dynHandler)

	t.Log("Adding keyvalue pairs to the external source")
	err := externalsource.AddKeyValue("testextkey1", "extkey1")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair externalsource")
	}
	err = externalsource.AddKeyValue("testextkey2", "extkey2")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair externalsource")
	}

	t.Log("verifying extsource configurations by GetConfigurations method")
	_, err = externalsource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get configurations from extsource")
	}

	t.Log("verifying extsource configurations by GetConfigurationByKey method")
	configkey1, err := externalsource.GetConfigurationByKey("testextkey1")
	if err != nil {
		t.Error("Failed to get config key from extsource")
	}

	//Accessing the extsource config key
	configkey2, err := externalsource.GetConfigurationByKey("testextkey2")
	if err != nil {
		t.Error("Failed to get config key from extsource")
	}

	if configkey1 != "extkey1" && configkey2 != "extkey2" {
		t.Error("extsource key value pairs is mismatched")
	}

	t.Log("Verifying the extsource priority")
	extsorcepriority := externalsource.GetPriority()
	if extsorcepriority != 4 {
		t.Error("extsource priority is mismatched")
	}

	t.Log("Verifying the extsource name")
	extsourcename := externalsource.GetSourceName()
	if extsourcename != "ExternalSource" {
		t.Error("extsource name is mismatched")
	}

	t.Log("verifying events")
	t.Log("create event")
	externalsource.AddKeyValue("testextkey3", "extkey3")
	t.Log("verifying created event")
	if dynHandler.EventKey != "testextkey3" && dynHandler.EventName != core.Create {
		t.Error("Failed to get the create event")
	}

	t.Log("update event")
	externalsource.AddKeyValue("testextkey3", "extkey33")
	t.Log("verifying update event")
	if dynHandler.EventKey != "testextkey3" && dynHandler.EventName != core.Update {
		t.Error("Failed to get the update event")
	}

	t.Log("envsource cleanup")
	extsourcecleanup := externalsource.Cleanup()
	if extsourcecleanup != nil {
		t.Error("extsource cleanup is Failed")
	}

	t.Log("verifying envsource configurations after cleanup")
	configkey1, _ = externalsource.GetConfigurationByKey("testextkey1")
	configkey2, _ = externalsource.GetConfigurationByKey("testextkey2")

	data, err := externalsource.GetConfigurationByKeyAndDimensionInfo("data@default#0.1", "hello")
	if data != nil || err != nil {
		t.Error("Failed to get configuration by dimension info and key")
	}

	if configkey1 != nil && configkey2 != nil {
		t.Error("envsource cleanup is Failed")
	}
}
