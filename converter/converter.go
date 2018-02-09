/*
 * Copyright 2017 Huawei Technologies Co., Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
* Created by on 2018/2/9.
 */

// Package converter provides a function to convert the provided data to yaml, json and java-properties
package converter

import (
	"errors"
	"fmt"

	"github.com/ServiceComb/go-chassis/core/lager"
	"gopkg.in/yaml.v2"
)

// Converter function is used convert provided data based on the format.
// Currently only yaml is supported.
func Converter(data []byte, format string) ([]byte, error) {
	var e error
	switch format {
	case "yaml":
		yamlContent, err := convertToYAML(data)
		if err != nil {
			return nil, err
		}
		return yamlContent, nil
	default:
		lager.Logger.Warnf(nil, "The supported format type is yaml but passed format is %s", format)
		err := fmt.Sprintf("The supported format type is yaml but passed format is %s", format)
		e = errors.New(err)
	}

	return nil, e
}

func convertToYAML(data []byte) ([]byte, error) {
	var object interface{}
	err := yaml.Unmarshal(data, &object)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(object)
}
