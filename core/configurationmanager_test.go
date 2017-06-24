package core

import (
	"testing"
)

type TestingSource struct {
	configuration  *map[string]interface{}
	changeCallback *ChangesCallback
	d              *Dispatcher
}

func (this *TestingSource) GetConfiguration() map[string]interface{} {
	return *this.configuration
}

func (this *TestingSource) AddDispatcher(dispatcher *Dispatcher) {
	this.d = dispatcher
}

func (this *TestingSource) GetPriority() int {
	return 0
}

func (this *TestingSource) GetSourceName() string {
	return "TestingSource"
}

func (this *TestingSource) AddDynamicConfigHandler(callback *ChangesCallback) {
	this.changeCallback = callback
}

func Test_AddSource(t *testing.T) {
	testConfig := &map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := &TestingSource{configuration: testConfig}

	cm := NewConfigurationManager(NewDispatcher())
	cm.AddSource(testSource)

	configInfo := cm.GetConfigurations()
	v := configInfo["aaa"]
	if v != "111" {
		t.Error("Error when pushing configuration from source to configurationmanager")
	}
	v = configInfo["bbb"]
	if v != "222" {
		t.Error("Error when pushing configuration from source to configurationmanager")
	}
}

func Test_Refresh(t *testing.T) {
	testConfig := &map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := &TestingSource{configuration: testConfig}

	cm := NewConfigurationManager(NewDispatcher())
	cm.AddSource(testSource)

	configInfo := cm.GetConfigurations()
	if len(configInfo) != 2 {
		t.Error("Config items error before refresh")
	}

	(*testConfig)["ccc"] = "333"
	if len(configInfo) != 2 {
		t.Error("Config items error before refresh after update source")
	}
	cm.Refresh()
	configInfo = cm.GetConfigurations()
	if len(configInfo) != 3 {
		t.Error("Config items error after refresh")
	}
}
