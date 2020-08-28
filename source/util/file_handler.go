package util

import (
	"fmt"
	"github.com/go-chassis/openlog"
	"gopkg.in/yaml.v2"
	"path/filepath"
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
		switch item.Value.(type) {
		case yaml.MapSlice:
			subResult := retrieveItems(prefix+item.Key.(string), item.Value.(yaml.MapSlice))
			for k, v := range subResult {
				result[k] = v
			}

		// Struct field to map
		case []interface{}:
			arr := make([]interface{}, 0)
			for _, v := range item.Value.([]interface{}) {
				if ms, ok := v.(yaml.MapSlice); ok {
					subResult := retrieveItems("", ms)
					arr = append(arr, subResult)
				}
			}

			k, ok := item.Key.(string)
			if !ok {
				openlog.Error("yaml path is not string", openlog.WithTags(
					openlog.Tags{
						"key": item.Key,
					},
				))
				continue
			}
			var keyVal = item.Value
			if val, ok := item.Value.(string); ok {
				keyVal = ExpandValueEnv(val)
			}
			result[prefix+k] = keyVal
			// replace  prefix+k with new arr value
			if len(arr) != 0 {
				result[prefix+k] = arr
			}

		default:
			k, ok := item.Key.(string)
			if !ok {
				openlog.Error("yaml tag is not string", openlog.WithTags(
					openlog.Tags{
						"key": item.Key,
					},
				))
				continue
			}

			var keyVal = item.Value
			if val, ok := item.Value.(string); ok {
				keyVal = ExpandValueEnv(val)
			}
			result[prefix+k] = keyVal

		}

	}

	return result
}

//UseFileNameAsKeyContentAsValue is a FileHandler, it sets the yaml file name as key and the content as value
func UseFileNameAsKeyContentAsValue(p string, content []byte) (map[string]interface{}, error) {
	_, filename := filepath.Split(p)
	configMap := make(map[string]interface{})
	configMap[filename] = content
	return configMap, nil
}

//Convert2configMap is legacy API
func Convert2configMap(p string, content []byte) (map[string]interface{}, error) {
	return UseFileNameAsKeyContentAsValue(p, content)
}
