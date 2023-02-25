/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package archaius

import "github.com/arielsrv/go-archaius/source"

var newFuncMap = make(map[string]NewRemoteSource)

const (
	// ApolloSource is for apollo source
	ApolloSource = "apollo"
	// ConfigCenterSource is for config center source
	ConfigCenterSource = "config-center"
	// KieSource is for ServiceComb-Kie source
	KieSource = "kie"
)

// NewRemoteSource create a new remote source
type NewRemoteSource func(info *RemoteInfo) (source.ConfigSource, error)

// InstallRemoteSource allow user customize remote source
func InstallRemoteSource(source string, remoteSource NewRemoteSource) {
	newFuncMap[source] = remoteSource
}
