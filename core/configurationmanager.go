package core

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

// Config manager Source
type ConfigMgr interface {
	AddSource(source ConfigurationSource) error
	Refresh()
	GetConfigurations() map[string]interface{}
	GetConfigurationsByKey(key string) interface{}
}

func (s ConfigSources) Len() int { return len(s) }

func (s ConfigSources) Less(i, j int) bool {
	return s[i].GetPriority() < s[j].GetPriority()
}

func (s ConfigSources) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ConfigurationManager struct {
	sources       ConfigSources
	configuration map[string]interface{}
	dispatcher    *Dispatcher
	sync.RWMutex
}

func (this *ConfigurationManager) AddSource(s ConfigurationSource) error {
	if s == nil || s.GetSourceName() == "" {
		return errors.New("nil or invalid source supplied")
	}
	this.Lock()
	var updateCallback ChangesCallback = this.updateHandler
	err := s.AddDynamicConfigHandler(&updateCallback)
	if err != nil {
		return err
	}
	this.sources = append(this.sources, s)
	sort.Sort(this.sources)
	this.Unlock()
	return nil
}

func (this *ConfigurationManager) GetConfigurations() map[string]interface{} {
	if this.configuration != nil && len(this.configuration) != 0 {
		return this.configuration
	} else {
		this.Refresh()
	}
	fmt.Printf("config is %v", this.configuration)
	return this.configuration
}

func (this *ConfigurationManager) Refresh() {
	this.RLock()
	defer this.RUnlock()

	this.configuration = make(map[string]interface{})
	for _, s := range this.sources {
		for k, v := range s.GetConfiguration() {
			this.configuration[k] = v
		}
	}
}

func (this *ConfigurationManager) GetConfigurationsByKey(key string) interface{} {
	sValues, ok := this.configuration[key]
	if !ok {
		return nil
	}

	return sValues
}

func (this *ConfigurationManager) updateHandler(event *Event) error {
	// refresh all configuration one by one
	if event == nil || event.EventSource == "" || event.EventName == "" {
		return errors.New("nil or invalid event supplied")
	}

	this.Lock()
	switch event.EventType {
	case DELETE:
		delete(this.configuration, event.EventName)
	case UPDATE:
		this.configuration[event.EventName] = event.Value
	case CREATE:
		this.configuration[event.EventName] = event.Value
	default:
		this.Unlock()
		return fmt.Errorf("the event type: %s do not support.", event.EventType)
	}
	this.Unlock()

	this.dispatcher.DispatchEvent(event)
	return nil
}
func NewConfigurationManager(d *Dispatcher) *ConfigurationManager {
	cm := ConfigurationManager{configuration: make(map[string]interface{}), dispatcher: d}
	for _, k := range DefaultSources {
		cm.AddSource(k)
	}
	return &cm
}
