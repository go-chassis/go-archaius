package goarchaius

import (
	"testing"
)

type TestingSource struct {
	configuration *map[string]interface{}
	d             *Dispatcher
}

func (this *TestingSource) GetConfiguration() map[string]interface{} {
	return *this.configuration
}

func (this *TestingSource) AddDispatcher(dispatcher *Dispatcher) {
	this.d = dispatcher
}

func Test_AddSource(t *testing.T) {
	testConfig := &map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := &TestingSource{configuration: testConfig}

	cm := NewConfigurationManager()
	cm.AddSource(testSource)

	configInfo := cm.Configuration
	v := configInfo["aaa"]
	if v != "111" {
		t.Error("Error when pushing configuration from source to configurationmanager")
	}
	v = configInfo["bbb"]
	if v != "222" {
		t.Error("Error when pushing configuration from source to configurationmanager")
	}
	ts := cm.Sources[0]
	if ts != testSource {
		t.Error("Error to recording source in ConfigManager")
	}
}

func Test_Refresh(t *testing.T) {
	testConfig := &map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := &TestingSource{configuration: testConfig}

	cm := NewConfigurationManager()
	cm.AddSource(testSource)

	configInfo := cm.Configuration
	if len(configInfo) != 2 {
		t.Error("Config items error before refresh")
	}

	(*testConfig)["ccc"] = "333"
	if len(configInfo) != 2 {
		t.Error("Config items error before refresh after update source")
	}
	cm.Refresh()
	configInfo = cm.Configuration
	if len(configInfo) != 3 {
		t.Error("Config items error after refresh")
	}
}
