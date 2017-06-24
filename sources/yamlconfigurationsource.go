package sources

import (
	"fmt"
	. "github.com/servicecomb/go-archaius/core"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type YamlConfigurationSource struct {
	configuration map[string]interface{}
	d             *Dispatcher
}

func (this *YamlConfigurationSource) GetConfiguration() map[string]interface{} {
	return this.configuration
}

func (this *YamlConfigurationSource) AddDispatcher(dispatcher *Dispatcher) {
	this.d = dispatcher
}

func NewYamlConfigurationSource(yamlFile string) (*YamlConfigurationSource, error) {
	//Read yaml file
	yamlContent, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}

	ss := yaml.MapSlice{}
	err = yaml.Unmarshal([]byte(yamlContent), &ss)
	config := retrieveItems("", ss)
	fmt.Printf("config is %s\n", config)
	source := &YamlConfigurationSource{configuration: config}

	return source, err
}

func (this *YamlConfigurationSource) GetPriority() int {
	return 12
}

func (this *YamlConfigurationSource) GetSourceName() string {
	return "YamlFile"
}

func retrieveItems(prefix string, subItems yaml.MapSlice) map[string]interface{} {
	fmt.Printf("prefix = %s\n", prefix)

	if prefix != "" {
		prefix += "."
	}

	result := map[string]interface{}{}

	for _, item := range subItems {
		//If there are sub-items existing
		_, isSlice := item.Value.(yaml.MapSlice)
		if isSlice {
			subResult := retrieveItems(prefix+item.Key.(string), item.Value.(yaml.MapSlice))
			for k, v := range subResult {
				result[k] = v
			}
		} else {
			result[prefix+item.Key.(string)] = item.Value
			fmt.Printf("config: %s = %s\n", prefix+item.Key.(string), item.Value)
		}
	}

	return result
}

func (this *YamlConfigurationSource) AddDynamicConfigHandler(callback *ChangesCallback) {
	return
}
