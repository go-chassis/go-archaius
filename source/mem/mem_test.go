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
package mem_test

import (
	"fmt"
	"testing"

	"github.com/arielsrv/go-archaius/event"
	"github.com/arielsrv/go-archaius/source/mem"
)

type TestDynamicConfigHandler struct {
	EventName  string
	EventKey   string
	EventValue interface{}
}

func (t *TestDynamicConfigHandler) OnEvent(e *event.Event) {
	t.EventKey = e.Key
	t.EventName = e.EventType
	t.EventValue = e.Value
}

func (t *TestDynamicConfigHandler) OnModuleEvent(events []*event.Event) {
	fmt.Println("implement me")
}

func TestMemoryConfigurationSource(t *testing.T) {

	memorysource := mem.NewMemoryConfigurationSource()

	t.Log("Test memorysource")

	dynHandler := new(TestDynamicConfigHandler)

	go memorysource.Watch(dynHandler)

	t.Log("Adding keyvalue pairs to the memory source")
	err := memorysource.Set("testextkey1", "extkey1")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair memorysource")
	}
	err = memorysource.Set("testextkey2", "extkey2")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair memorysource")
	}

	err = memorysource.Set("testmemkey2", "memkey2")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair memorysource")
	}

	err = memorysource.Delete("testmemkey2")
	if err != nil {
		t.Error("Failed to Add Keyvalue pair memorysource")
	}

	t.Log("verifying memorysource configurations by Configs method")
	_, err = memorysource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get configurations from extsource")
	}

	t.Log("verifying memorysource configurations by GetConfigurationByKey method")
	configkey1, err := memorysource.GetConfigurationByKey("testextkey1")
	if err != nil {
		t.Error("Failed to get config key from extsource")
	}

	//Accessing the extsource config key
	configkey2, err := memorysource.GetConfigurationByKey("testextkey2")
	if err != nil {
		t.Error("Failed to get config key from extsource")
	}

	if configkey1 != "extkey1" && configkey2 != "extkey2" {
		t.Error("memorysource key value pairs is mismatched")
	}

	t.Log("Verifying the memorysource priority")
	memsorcepriority := memorysource.GetPriority()
	if memsorcepriority != 1 {
		t.Error("memorysource priority is mismatched")
	}

	t.Log("Verifying the memorysource name")
	memsourcename := memorysource.GetSourceName()
	if memsourcename != "MemorySource" {
		t.Error("memorysource name is mismatched")
	}

	t.Log("verifying events")
	t.Log("create event")
	memorysource.Set("testextkey3", "extkey3")
	t.Log("verifying created event")
	if dynHandler.EventKey != "testextkey3" && dynHandler.EventName != event.Create {
		t.Error("Failed to get the create event")
	}

	t.Log("update event")
	memorysource.Set("testextkey3", "extkey33")
	t.Log("verifying update event")
	if dynHandler.EventKey != "testextkey3" && dynHandler.EventName != event.Update {
		t.Error("Failed to get the update event")
	}

	memorysource.Set("testmemkey3", "memkey33")
	t.Log("verifying update event")
	if dynHandler.EventKey != "testmemkey3" && dynHandler.EventName != event.Update {
		t.Error("Failed to get the update event")
	}

	err = memorysource.Delete("testmemkey3")
	if err != nil {
		t.Error("Failed to delete Keyvalue pair memorysource")
	}

	t.Log("memorysource cleanup")
	extsourcecleanup := memorysource.Cleanup()
	if extsourcecleanup != nil {
		t.Error("memorysource cleanup is Failed")
	}

	t.Log("verifying envsource configurations after cleanup")
	configkey1, _ = memorysource.GetConfigurationByKey("testextkey1")
	configkey2, _ = memorysource.GetConfigurationByKey("testextkey2")

	if configkey1 != nil && configkey2 != nil {
		t.Error("memorysource cleanup is Failed")
	}
}
