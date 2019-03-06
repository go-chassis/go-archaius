// Package archaius provides you APIs which helps to manage files,
// remote config center configurations
package archaius

import (
	"crypto/tls"
	"os"
	"strings"
	"sync"

	"errors"
	"github.com/go-chassis/go-archaius/core"
	"github.com/go-chassis/go-archaius/sources/configcenter"
	"github.com/go-chassis/go-archaius/sources/file-source"
	"github.com/go-chassis/go-archaius/sources/memory-source"
	"github.com/go-chassis/go-cc-client"
	"github.com/go-mesh/openlogging"
)

var (
	factory ConfigurationFactory
	fs      filesource.FileSource
	ms      = memoryconfigsource.NewMemoryConfigurationSource()

	once             = sync.Once{}
	onceConfigCenter = sync.Once{}
	onceExternal     = sync.Once{}
)

// ConfigCenterInfo has attribute for config center source initialization
type ConfigCenterInfo struct {
	//required.
	//Key value can be in different namespace, we call it dimension.
	//although key is same but in different dimension, the value is different.
	//you must specify it, so that the config center source will just pull this dimension's key value
	DimensionInfo string

	//archaius config center source support 2 types of refresh mechanism:
	//0: Web-Socket Based -  client makes an web socket connection with
	//the config server and keeps getting an events whenever any data changes.
	//1: Pull Configuration interval- In this type client keeps polling the configuration from
	//the config server at regular intervals.
	RefreshMode int

	//Pull Configuration interval, unit is second
	RefreshInterval int

	//Configurations for config client implementation
	//if you alread create a client, don't need to set those config
	URL           string
	TenantName    string
	EnableSSL     bool
	TLSConfig     *tls.Config
	AutoDiscovery bool
	ClientType    string
	Version       string
	RefreshPort   string
	Environment   string
}

func initFileSource(o *Options) (core.ConfigSource, error) {
	files := make([]string, 0)
	// created file source object
	fs = filesource.NewFileSource()
	// adding all files with file source
	for _, v := range o.RequiredFiles {
		if err := fs.AddFile(v, filesource.DefaultFilePriority, o.FileHandler); err != nil {
			openlogging.GetLogger().Errorf("add file source error [%s].", err.Error())
			return nil, err
		}
		files = append(files, v)
	}
	for _, v := range o.OptionalFiles {
		_, err := os.Stat(v)
		if os.IsNotExist(err) {
			openlogging.GetLogger().Infof("[%s] not exist", v)
			continue
		}
		if err := fs.AddFile(v, filesource.DefaultFilePriority, o.FileHandler); err != nil {
			openlogging.GetLogger().Infof("%v", err)
			return nil, err
		}
		files = append(files, v)
	}
	openlogging.GetLogger().Infof("Configuration files: %s", strings.Join(files, ", "))
	return fs, nil
}

// Init create a Archaius config singleton
func Init(opts ...Option) error {
	var errG error
	once.Do(func() {
		var err error
		o := &Options{}
		for _, opt := range opts {
			opt(o)
		}

		// created config factory object
		factory, err = NewConfigFactory()
		if err != nil {
			errG = err
			return
		}
		factory.DeInit()
		factory.Init()

		fs, err := initFileSource(o)
		if err != nil {
			errG = err
			return
		}
		if o.ConfigCenterInfo != (ConfigCenterInfo{}) {
			if err := InitConfigCenterSource(o.ConfigCenterInfo, o.ConfigClient); err != nil {
				errG = err
				return
			}
		}
		err = factory.AddSource(fs)
		if err != nil {
			errG = err
			return
		}
		eventHandler := EventListener{
			Name:    "EventHandler",
			Factory: factory,
		}

		factory.RegisterListener(eventHandler, "a*")
		openlogging.GetLogger().Info("archaius init success")
	})

	return errG
}

//InitConfigCenterSource create a config center source singleton
//A config center source pull remote config server key values into local memory
//so that you can use GetXXX to get value easily
func InitConfigCenterSource(ci ConfigCenterInfo, cc ccclient.ConfigClient) error {
	var errG error
	if ci == (ConfigCenterInfo{}) {
		return errors.New("ConfigCenterInfo can not be empty")
	}
	onceConfigCenter.Do(func() {
		var err error
		if cc == nil {
			opts := ccclient.Options{
				DimensionInfo: ci.DimensionInfo,
				ServerURI:     ci.URL,
				TenantName:    ci.TenantName,
				EnableSSL:     ci.EnableSSL,
				TLSConfig:     ci.TLSConfig,
				RefreshPort:   ci.RefreshPort,
				AutoDiscovery: ci.AutoDiscovery,
				Env:           ci.Environment,
			}
			cc, err = ccclient.NewClient(ci.ClientType, opts)
			if err != nil {
				errG = err
				return
			}
		}
		configCenterSource := configcenter.NewConfigCenterSource(cc,
			ci.DimensionInfo, ci.RefreshMode,
			ci.RefreshInterval)
		err = factory.AddSource(configCenterSource)
		if err != nil {
			errG = err
			return
		}

		eventHandler := EventListener{
			Name:    "EventHandler",
			Factory: factory,
		}

		factory.RegisterListener(eventHandler, "a*")
	})

	return errG
}

// InitExternal create any config singleton
func InitExternal(opts ...Option) error {
	var errG error
	onceExternal.Do(func() {
		var err error
		o := &Options{}
		for _, opt := range opts {
			opt(o)
		}

		factory, err = NewConfigFactory()
		if err != nil {
			errG = err
			return
		}

		factory.DeInit()
		factory.Init()

		err = factory.AddSource(o.ExternalSource)
		if err != nil {
			errG = err
			return
		}

		eventHandler := EventListener{
			Name:    "EventHandler",
			Factory: factory,
		}

		factory.RegisterListener(eventHandler, "a*")

	})

	return errG
}

// EventListener is a struct having information about registering key and object
type EventListener struct {
	Name    string
	Factory ConfigurationFactory
}

// Event is invoked while generating events at run time
func (e EventListener) Event(event *core.Event) {
	value := e.Factory.GetConfigurationByKey(event.Key)
	openlogging.GetLogger().Infof("config value after change %s | %s", event.Key, value)
}

// Get is for to get the value of configuration key
func Get(key string) interface{} {
	return factory.GetConfigurationByKey(key)
}

// Exist check the configuration key existence
func Exist(key string) bool {
	return factory.IsKeyExist(key)
}

// UnmarshalConfig is for unmarshalling the configuraions of receiving object
func UnmarshalConfig(obj interface{}) error {
	return factory.Unmarshal(obj)
}

// GetBool is gives the key value in the form of bool
func GetBool(key string, defaultValue bool) bool {
	b, err := factory.GetValue(key).ToBool()
	if err != nil {
		return defaultValue
	}
	return b
}

// GetFloat64 gives the key value in the form of float64
func GetFloat64(key string, defaultValue float64) float64 {
	result, err := factory.GetValue(key).ToFloat64()
	if err != nil {
		return defaultValue
	}
	return result
}

// GetInt gives the key value in the form of GetInt
func GetInt(key string, defaultValue int) int {
	result, err := factory.GetValue(key).ToInt()
	if err != nil {
		return defaultValue
	}
	return result
}

// GetString gives the key value in the form of GetString
func GetString(key string, defaultValue string) string {
	result, err := factory.GetValue(key).ToString()
	if err != nil {
		return defaultValue
	}
	return result
}

// GetConfigs gives the information about all configurations
func GetConfigs() map[string]interface{} {
	return factory.GetConfigurations()
}

// GetStringByDI is for to get the value of configuration key based on dimension info
func GetStringByDI(dimensionInfo, key string, defaultValue string) string {
	result, err := factory.GetValueByDI(dimensionInfo, key).ToString()
	if err != nil {
		return defaultValue
	}
	return result
}

// GetConfigsByDI is for to get the all configurations received dimensionInfo
func GetConfigsByDI(dimensionInfo string) map[string]interface{} {
	return factory.GetConfigurationsByDimensionInfo(dimensionInfo)
}

// AddDI adds a NewDimensionInfo of which configurations needs to be taken
func AddDI(dimensionInfo string) (map[string]string, error) {
	config, err := factory.AddByDimensionInfo(dimensionInfo)
	return config, err
}

//RegisterListener to Register all listener for different key changes, each key could be a regular expression
func RegisterListener(listenerObj core.EventListener, key ...string) error {
	return factory.RegisterListener(listenerObj, key...)
}

// UnRegisterListener is to remove the listener
func UnRegisterListener(listenerObj core.EventListener, key ...string) error {
	return factory.UnRegisterListener(listenerObj, key...)
}

// AddFile is for to add the configuration files into the configfactory at run time
func AddFile(file string, opts ...FileOption) error {
	o := &FileOptions{}
	for _, f := range opts {
		f(o)
	}
	if err := fs.AddFile(file, filesource.DefaultFilePriority, o.Handler); err != nil {
		return err
	}
	return factory.Refresh(fs.GetSourceName())
}

//AddKeyValue is for to add the configuration key, value pairs into the configfactory at run time
// it is just affect the local configs
func AddKeyValue(key string, value interface{}) error {
	return ms.AddKeyValue(key, value)
}

// DeleteKeyValue is for to delete the configuration key, value pairs into the configfactory at run time
func DeleteKeyValue(key string, value interface{}) error {
	return ms.DeleteKeyValue(key, value)
}

//AddSource add source implementation
func AddSource(source core.ConfigSource) error {
	return factory.AddSource(source)
}

//GetConfigFactory return factory
func GetConfigFactory() ConfigurationFactory {
	return factory
}
