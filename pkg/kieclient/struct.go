/*
 * Copyright 2020 Huawei Technologies Co., Ltd
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

package kieclient

//KVRequest is http request body
type KVRequest struct {
	ID        string            `json:"id" yaml:"id"`
	Key       string            `json:"key" yaml:"key"`
	Value     string            `json:"value,omitempty" yaml:"value,omitempty"`
	Status    string            `json:"status,omitempty" yaml:"status,omitempty"`
	ValueType string            `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty"` //ini,json,text,yaml,properties
	Checker   string            `json:"check,omitempty" yaml:"check,omitempty"`                                       //python script
	Labels    map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`                                     //redundant
}

//KVResponse represents the key value list
type KVResponse struct {
	LabelDoc *LabelDocResponse `json:"label,omitempty"`
	Total    int               `json:"total,omitempty"`
	Data     []*KVDoc          `json:"data,omitempty"`
}

//LabelDocResponse is label struct
type LabelDocResponse struct {
	Labels map[string]string `json:"labels,omitempty"`
}

//KVDoc is database struct to store kv
type KVDoc struct {
	ID             string `json:"id,omitempty" bson:"id,omitempty" yaml:"id,omitempty" swag:"string"`
	Key            string `json:"key" yaml:"key"`
	Value          string `json:"value,omitempty" yaml:"value,omitempty"`
	ValueType      string `json:"value_type,omitempty" bson:"value_type,omitempty" yaml:"value_type,omitempty"` //ini,json,text,yaml,properties
	Checker        string `json:"check,omitempty" yaml:"check,omitempty"`                                       //python script
	CreateRevision int64  `json:"create_revision,omitempty" bson:"create_revision," yaml:"create_revision,omitempty"`
	UpdateRevision int64  `json:"update_revision,omitempty" bson:"update_revision," yaml:"update_revision,omitempty"`
	Project        string `json:"project,omitempty" yaml:"project,omitempty"`
	Status         string `json:"status,omitempty" yaml:"status,omitempty"`
	CreatTime      int64  `json:"create_time,omitempty" yaml:"create_time,omitempty"`
	UpdateTime     int64  `json:"update_time,omitempty" yaml:"update_time,omitempty"`

	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"` //redundant
	Domain string            `json:"domain,omitempty" yaml:"domain,omitempty"` //redundant

}
