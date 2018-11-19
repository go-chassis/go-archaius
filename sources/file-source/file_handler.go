package filesource

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

//FileHandler decide how to convert a file content into key values
//archaius will manage file content as those key values
type FileHandler func(filePath string, content []byte) (map[string]interface{}, error)

//Convert2JavaProps is a FileHandler
//it convert the yaml content into java props
func Convert2JavaProps(p string, content []byte) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	ss := yaml.MapSlice{}
	err := yaml.Unmarshal([]byte(content), &ss)
	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal [%s] failed, %s", content, err)
	}
	configMap = retrieveItems("", ss)

	return configMap, nil
}
func retrieveItems(prefix string, subItems yaml.MapSlice) map[string]interface{} {
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
		}
	}

	return result
}

//Convert2configMap  converts the yaml file file name as key and the content as value
func Convert2configMap(p string, content []byte) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})
	configMap[p] = content
	return configMap, nil
}
