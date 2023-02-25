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

package configcenter

import (
	"crypto/tls"
	"net/http"
)

// Options  is remote client option
type Options struct {
	DefaultDimension string
	Service          string
	App              string
	Version          string
	Env              string

	ConfigServerAddresses []string
	RefreshPort           string
	APIVersion            string
	TLSConfig             *tls.Config
	TenantName            string
	EnableSSL             bool
}

// GetDefaultHeaders gets default headers
func GetDefaultHeaders(tenantName string) http.Header {
	headers := http.Header{
		HeaderContentType: []string{"application/json"},
		HeaderUserAgent:   []string{"cse-configcenter-client/1.0.0"},
		HeaderTenantName:  []string{tenantName},
	}
	if environmentConfig != "" {
		headers.Set(HeaderEnvironment, environmentConfig)
	}

	return headers
}
