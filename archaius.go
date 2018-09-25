// Package archaius provides you APIs which helps to manage files,
// remote config center configurations
package archaius

import (
	"os"
	"strings"

	"github.com/go-chassis/go-archaius/core"
	"github.com/go-chassis/go-archaius/sources/file-source"
	"github.com/go-chassis/go-archaius/sources/memory-source"
	"github.com/go-mesh/openlogging"
	"sync"
)

var factory ConfigurationFactory
var fs filesource.FileSource
var memorySource memoryconfigsource.MemorySource

//ConfigCenterInfo holds config center info
//TODO add more infos
type ConfigCenterInfo struct {
	URL string
}

func initFileSource(o *Options) (core.ConfigSource, error) {
	files := make([]string, 0)
	// created file source object
	fs = filesource.NewYamlConfigurationSource()
	// adding all files with file source
	for _, v := range o.RequiredFiles {
		if err := fs.AddFileSource(v, filesource.DefaultFilePriority); err != nil {
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
		if err := fs.AddFileSource(v, filesource.DefaultFilePriority); err != nil {
			openlogging.GetLogger().Infof("%v", err)
			return nil, err
		}
		files = append(files, v)
	}
	openlogging.GetLogger().Infof("Configuration files: %s", strings.Join(files, ", "))
	return fs, nil
}

var once = sync.Once{}

// Init create a Archaius config singleton
// TODO init logic for config source
func Init(opts ...Option) error {
	var errG error
	once.Do(func() {
		var err error
		o := &Options{}
		for _, opt := range opts {
			opt(o)
		}

		// created config factory object
		factory, err = NewConfigFactory(openlogging.GetLogger())
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
		err = factory.AddSource(fs)
		if err != nil {
			errG = err
			return
		}
		for _, l := range o.EventListeners {
			factory.RegisterListener(l, l.Keys()...)

		}

	})

	return errG
}

// Get is for to get the value of configuration key
func Get(key string) interface{} {
	return factory.GetConfigurationByKey(key)
}

// Exist is check the configuration key existence
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
func AddFile(file string) error {
	return fs.AddFileSource(file, filesource.DefaultFilePriority)
}

//AddKeyValue is for to add the configuration key, value pairs into the configfactory at run time
// it is just affect the local configs
func AddKeyValue(key string, value interface{}) error {
	return memorySource.AddKeyValue(key, value)
}

// DeleteKeyValue is for to delete the configuration key, value pairs into the configfactory at run time
func DeleteKeyValue(key string, value interface{}) error {
	return memorySource.DeleteKeyValue(key, value)
}
