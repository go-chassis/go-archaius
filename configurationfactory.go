package goarchaius

import (
	"github.com/servicecomb/go-archaius/core"
	_ "github.com/servicecomb/go-archaius/sources"
)

type ConfigurationFactory struct {
	configMgr   core.ConfigMgr
	dispatcher  *core.Dispatcher
	initSuccess bool
}

var con *ConfigurationFactory

func NewConfigurationFactory() *ConfigurationFactory {
	if con == nil {
		con = new(ConfigurationFactory)
		con.dispatcher = core.NewDispatcher()
		con.configMgr = core.NewConfigurationManager(con.dispatcher)
	}
	return con
}

func (this *ConfigurationFactory) Init() error {
	this.configMgr.Refresh()
	this.initSuccess = true
	return nil
}

func (this *ConfigurationFactory) GetConfigurations() map[string]interface{} {
	if this.initSuccess == false {
		return nil
	}

	return this.configMgr.GetConfigurations()
}

// return all values of different sources
func (this *ConfigurationFactory) GetConfigurationByKey(key string) interface{} {
	if this.initSuccess == false {
		return nil
	}

	return this.configMgr.GetConfigurationsByKey(key)
}

//Function to Register all listener for different key changes
func (this *ConfigurationFactory) RegisterListener(keyName string, listenerObj *core.EventCallback) {
	this.dispatcher.AddEventListener(keyName, listenerObj)
}

// remove listener
func (this *ConfigurationFactory) RemoveListener(keyName string, listenerObj *core.EventCallback) {
	this.dispatcher.RemoveEventListener(keyName, listenerObj)
}
