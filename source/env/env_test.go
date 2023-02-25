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
package env_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/arielsrv/go-archaius/event"
	"github.com/arielsrv/go-archaius/source/env"
	"github.com/stretchr/testify/assert"
)

type TestDynamicConfigHandler struct{}

func (t *TestDynamicConfigHandler) OnModuleEvent(events []*event.Event) {
	fmt.Println("implement me")
}

func (t *TestDynamicConfigHandler) OnEvent(e *event.Event) {}

func populatEnvConfiguration() {

	os.Setenv("testenvkey1", "envkey1")
	os.Setenv("testenvkey2", "envkey2")
	os.Setenv("testenvkey3", "a=b=c")

}

func TestEnvConfigurationSource(t *testing.T) {

	populatEnvConfiguration()
	envsource := env.NewEnvConfigurationSource()
	t.Run("set env with underscore, use dot to get ", func(t *testing.T) {
		os.Setenv("a_b_c_d", "asd")
		envsource := env.NewEnvConfigurationSource()
		v, err := envsource.GetConfigurationByKey("a.b.c.d")
		assert.Equal(t, nil, err)
		assert.Equal(t, "asd", v)
		v, err = envsource.GetConfigurationByKey("a_b_c_d")
		assert.Equal(t, nil, err)
		assert.Equal(t, "asd", v)
	})
	t.Log("Test envconfigurationsource.go")

	t.Log("verifying envsource configurations by Configs method")
	_, err := envsource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get configurations from envsource")
	}

	t.Log("verifying envsource configurations by GetConfigurationByKey method")
	configkey1, err := envsource.GetConfigurationByKey("testenvkey1")
	if err != nil {
		t.Error("Failed to get existing configuration key value pair from envsource")
	}

	//Accessing the envsource config key
	configkey2, err := envsource.GetConfigurationByKey("testenvkey3")
	if err != nil {
		t.Error("Failed to get existing configuration key value pair from envsource")
	}

	if configkey1 != "envkey1" && configkey2 != "a=b=c" {
		t.Error("envsource key value pairs is mismatched")
	}

	t.Log("Verifying the envsource priority")
	envsorcepriority := envsource.GetPriority()
	if envsorcepriority != 3 {
		t.Error("envsource priority is mismatched")
	}

	t.Log("Verifying the envsource name")
	envsourcename := envsource.GetSourceName()
	if envsourcename != "EnvironmentSource" {
		t.Error("envsource name is mismatched")
	}

	dynHandler := new(TestDynamicConfigHandler)
	envdynamicconfig := envsource.Watch(dynHandler)
	if envdynamicconfig != nil {
		t.Error("Failed to get envsource dynamic configuration")
	}

	t.Log("envsource cleanup")
	envsourcecleanup := envsource.Cleanup()
	if envsourcecleanup != nil {
		t.Error("envsource cleanup is Failed")
	}

	t.Log("verifying envsource configurations after cleanup")
	configkey1, _ = envsource.GetConfigurationByKey("testenvkey1")
	configkey2, _ = envsource.GetConfigurationByKey("testenvkey2")
	if configkey1 != nil && configkey2 != nil {
		t.Error("envsource cleanup is Failed")
	}
}
