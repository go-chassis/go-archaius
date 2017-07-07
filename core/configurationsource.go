package core

type ChangesCallback func(*Event) error

// Config source should implement this interface
type ConfigurationSource interface {
	GetPriority() int
	GetSourceName() string
	GetConfiguration() map[string]interface{}
	AddDynamicConfigHandler(callback *ChangesCallback) error
}

type ConfigSources []ConfigurationSource

var DefaultSources ConfigSources = []ConfigurationSource{}
