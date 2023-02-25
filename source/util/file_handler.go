package util

import (
	"fmt"
	"path/filepath"

	"github.com/go-chassis/openlog"
	"gopkg.in/yaml.v2"
)

// FileHandler decide how to convert a file content into key values
// archaius will manage file content as those key values
type FileHandler func(filePath string, content []byte) (map[string]interface{}, error)

// Convert2JavaProps is a FileHandler
// it convert the yaml content into java props
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
		//check the item key first
		k, ok := checkKey(item.Key)
		if !ok {
			continue
		}
		//If there are sub-items existing
		switch item.Value.(type) {
		//sub items in a map
		case yaml.MapSlice:
			subResult := retrieveItems(prefix+item.Key.(string), item.Value.(yaml.MapSlice))
			for k, v := range subResult {
				result[k] = v
			}

		// sub items in an array
		case []interface{}:
			keyVal := item.Value.([]interface{})
			result[prefix+k] = retrieveItemInSlice(keyVal)

		// sub item is a string
		case string:
			result[prefix+k] = ExpandValueEnv(item.Value.(string))

		// sub item in other type
		default:
			result[prefix+k] = item.Value

		}

	}

	return result
}

func checkKey(key interface{}) (string, bool) {
	k, ok := key.(string)
	if !ok {
		openlog.Error("yaml tag is not string", openlog.WithTags(
			openlog.Tags{
				"key": key,
			},
		))
		return "", false
	}
	return k, true
}

func retrieveItemInSlice(value []interface{}) []interface{} {
	for i, v := range value {
		switch v.(type) {
		case yaml.MapSlice:
			value[i] = retrieveItems("", v.(yaml.MapSlice))
		case string:
			value[i] = ExpandValueEnv(v.(string))
		default:
			//do nothing
		}
	}
	return value
}

// UseFileNameAsKeyContentAsValue is a FileHandler, it sets the yaml file name as key and the content as value
func UseFileNameAsKeyContentAsValue(p string, content []byte) (map[string]interface{}, error) {
	_, filename := filepath.Split(p)
	configMap := make(map[string]interface{})
	configMap[filename] = content
	return configMap, nil
}

// Convert2configMap is legacy API
func Convert2configMap(p string, content []byte) (map[string]interface{}, error) {
	return UseFileNameAsKeyContentAsValue(p, content)
}
