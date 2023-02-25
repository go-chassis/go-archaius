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
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/arielsrv/go-archaius/pkg/serializers"
	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/openlog"
	"github.com/gorilla/websocket"
)

const (
	defaultTimeout = 10 * time.Second
	numberSign     = "%23"
	//StatusUP is a variable of type string
	StatusUP = "UP"
	//HeaderContentType is a variable of type string
	HeaderContentType = "Content-Type"
	//HeaderUserAgent is a variable of type string
	HeaderUserAgent = "User-Agent"
	//HeaderEnvironment specifies the environment of a service
	HeaderEnvironment        = "X-Environment"
	members                  = "/configuration/members"
	dimensionsInfo           = "dimensionsInfo"
	dynamicConfigAPI         = `/configuration/refresh/items`
	getConfigAPI             = `/configuration/items`
	defaultContentType       = "application/json"
	envProjectID             = "CSE_PROJECT_ID"
	packageInitError         = "package not initialize successfully"
	emptyConfigServerMembers = "empty config server member"
	emptyConfigServerConfig  = "empty config server passed"
	// Name of the Plugin
	Name = "config_center"
)

var (
	//HeaderTenantName is a variable of type string
	HeaderTenantName = "X-Tenant-Name"
	//ConfigMembersPath is a variable of type string
	ConfigMembersPath = ""
	//ConfigPath is a variable of type string
	ConfigPath = ""
	//ConfigRefreshPath is a variable of type string
	ConfigRefreshPath = ""
	autoDiscoverable  = false
	apiVersionConfig  = ""
	environmentConfig = ""
)

// Client is a struct
type Client struct {
	opts Options
	sync.RWMutex
	c            *httpclient.Requests
	wsDialer     *websocket.Dialer
	wsConnection *websocket.Conn
}

// New create cc client
func New(opts Options) (*Client, error) {
	var apiVersion string
	apiVersionConfig = opts.APIVersion
	switch apiVersionConfig {
	case "v2":
		apiVersion = "v2"
	case "V2":
		apiVersion = "v2"
	case "v3":
		apiVersion = "v3"
	case "V3":
		apiVersion = "v3"
	default:
		apiVersion = "v3"
	}
	updateAPIPath(apiVersion)

	hc, err := httpclient.New(&httpclient.Options{
		TLSConfig:  opts.TLSConfig,
		Compressed: false,
	})
	if err != nil {
		return nil, err
	}
	c := &Client{
		c:    hc,
		opts: opts,
		wsDialer: &websocket.Dialer{
			TLSClientConfig:  opts.TLSConfig,
			HandshakeTimeout: defaultTimeout,
		},
	}
	c.Shuffle()
	return c, nil
}

// Update the Base PATH and HEADERS Based on the version of ConfigCenter used.
func updateAPIPath(apiVersion string) {
	//Check for the env Name in Container to get Domain Name
	//Default value is  "default"
	projectID, isExist := os.LookupEnv(envProjectID)
	if !isExist {
		projectID = "default"
	}
	switch apiVersion {
	case "v3":
		ConfigMembersPath = "/v3/" + projectID + members
		ConfigPath = "/v3/" + projectID + getConfigAPI
		ConfigRefreshPath = "/v3/" + projectID + dynamicConfigAPI
	case "v2":
		ConfigMembersPath = "/members"
		ConfigPath = "/configuration/v2/items"
		ConfigRefreshPath = "/configuration/v2/refresh/items"
	default:
		ConfigMembersPath = "/v3/" + projectID + members
		ConfigPath = "/v3/" + projectID + getConfigAPI
		ConfigRefreshPath = "/v3/" + projectID + dynamicConfigAPI
	}
}

func (c *Client) call(method string, api string, headers http.Header, body []byte, s interface{}) error {
	hosts, err := c.GetConfigServer()
	if err != nil {
		openlog.Error("Get config server addr failed:" + err.Error())
	}
	index := rand.Int() % len(c.opts.ConfigServerAddresses)
	host := hosts[index]
	rawURI := host + api
	errMsgPrefix := fmt.Sprintf("Call %s failed: ", rawURI)
	resp, err := c.HTTPDo(method, rawURI, headers, body)
	if err != nil {
		openlog.Error(errMsgPrefix + err.Error())
		return err

	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		openlog.Error(errMsgPrefix + err.Error())
		return err
	}
	if !isStatusSuccess(resp.StatusCode) {
		err = fmt.Errorf("statusCode: %d, resp body: %s", resp.StatusCode, body)
		openlog.Error(errMsgPrefix + err.Error())
		return err
	}
	contentType := resp.Header.Get("Content-Type")
	if len(contentType) > 0 && (len(defaultContentType) > 0 && !strings.Contains(contentType, defaultContentType)) {
		err = fmt.Errorf("content type not %s", defaultContentType)
		openlog.Error(errMsgPrefix + err.Error())
		return err
	}
	err = serializers.Decode(defaultContentType, body, s)
	if err != nil {
		openlog.Error("Decode failed:" + err.Error())
		return err
	}
	return nil
}

// HTTPDo Use http-client package for rest communication
func (c *Client) HTTPDo(method string, rawURL string, headers http.Header, body []byte) (resp *http.Response, err error) {
	if len(headers) == 0 {
		headers = make(http.Header)
	}
	for k, v := range GetDefaultHeaders(c.opts.TenantName) {
		headers[k] = v
	}
	return c.c.Do(context.Background(), method, rawURL, headers, body)
}

// Flatten pulls all the configuration from config center and merge kv in different dimension
func (c *Client) Flatten(dimensionInfo string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	configAPIResp, err := c.PullGroupByDimension(dimensionInfo)
	if err != nil {
		openlog.Error("Flatten config failed:" + err.Error())
		return nil, err
	}
	for _, v := range configAPIResp {
		for key, value := range v {
			config[key] = value

		}
	}
	return config, nil
}

// PullGroupByDimension pulls all the configuration from Config-Server group by dimesion Info
func (c *Client) PullGroupByDimension(dimensionInfo string) (map[string]map[string]interface{}, error) {
	configAPIRes := make(map[string]map[string]interface{})
	parsedDimensionInfo := strings.Replace(dimensionInfo, "#", numberSign, -1)
	restAPI := ConfigPath + "?" + dimensionsInfo + "=" + parsedDimensionInfo
	err := c.call(http.MethodGet, restAPI, nil, nil, &configAPIRes)
	if err != nil {
		openlog.Error("Flatten config failed:" + err.Error())
		return nil, err
	}

	return configAPIRes, nil
}

// Do is common http remote call
func (c *Client) Do(method string, data interface{}) (map[string]interface{}, error) {
	configAPIS := make(map[string]interface{})
	body, err := serializers.Encode(serializers.JSONEncoder, data)
	if err != nil {
		openlog.Error("serializer data failed , err :" + err.Error())
		return nil, err
	}
	err = c.call(method, ConfigPath, nil, body, &configAPIS)
	if err != nil {
		return nil, err
	}
	return configAPIS, nil
}

// AddConfig post new config
func (c *Client) AddConfig(data *CreateConfigAPI) (map[string]interface{}, error) {
	return c.Do("POST", data)
}

// DeleteConfig delete configs
func (c *Client) DeleteConfig(data *DeleteConfigAPI) (map[string]interface{}, error) {
	return c.Do("DELETE", data)
}

// Watch use websocket
func (c *Client) Watch(f func(map[string]interface{}), errHandler func(err error)) error {
	parsedDimensionInfo := strings.Replace(c.opts.DefaultDimension, "#", numberSign, -1)
	refreshConfigPath := ConfigRefreshPath + `?` + dimensionsInfo + `=` + parsedDimensionInfo
	if c.wsDialer != nil {
		/*-----------------
		1. Decide on the URL
		2. Create WebSocket Connection
		3. Call KeepAlive in separate thread
		3. Generate events on Receive Data
		*/
		baseURL, err := c.getWebSocketURL()
		if err != nil {
			error := errors.New("error in getting default server info")
			return error
		}
		url := baseURL.String() + refreshConfigPath
		c.wsConnection, _, err = c.wsDialer.Dial(url, nil)
		if err != nil {
			return fmt.Errorf("watching config-center dial catch an exception error:%s", err.Error())
		}
		keepAlive(c.wsConnection, 15*time.Second)
		go func() error {
			for {
				messageType, message, err := c.wsConnection.ReadMessage()
				if err != nil {
					break
				}
				if messageType == websocket.TextMessage {
					m, err := GetConfigs(message)
					if err != nil {
						errHandler(err)
						continue
					}
					f(m)
				}
			}
			err = c.wsConnection.Close()
			if err != nil {
				openlog.Error(err.Error())
				return fmt.Errorf("CC watch Conn close failed error:%s", err.Error())
			}
			return nil
		}()
	}
	return nil
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})
	go func() {
		for {
			err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				return
			}
			time.Sleep(timeout / 2)
			if time.Now().Sub(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}

func isStatusSuccess(i int) bool {
	return i >= http.StatusOK && i < http.StatusBadRequest
}

// Shuffle is a method to log error
func (c *Client) Shuffle() error {
	if c.opts.ConfigServerAddresses == nil || len(c.opts.ConfigServerAddresses) == 0 {
		err := errors.New(emptyConfigServerConfig)
		openlog.Error(emptyConfigServerConfig)
		return err
	}

	perm := rand.Perm(len(c.opts.ConfigServerAddresses))

	c.Lock()
	defer c.Unlock()
	openlog.Debug(fmt.Sprintf("before shuffled member %s ", c.opts.ConfigServerAddresses))
	for i, v := range perm {
		openlog.Debug(fmt.Sprintf("shuffler %d %d", i, v))
		tmp := c.opts.ConfigServerAddresses[v]
		c.opts.ConfigServerAddresses[v] = c.opts.ConfigServerAddresses[i]
		c.opts.ConfigServerAddresses[i] = tmp
	}

	openlog.Debug(fmt.Sprintf("shuffled member %s", c.opts.ConfigServerAddresses))
	return nil
}

// GetConfigServer is a method used for getting server configuration
func (c *Client) GetConfigServer() ([]string, error) {

	if len(c.opts.ConfigServerAddresses) == 0 {
		err := errors.New(emptyConfigServerMembers)
		openlog.Error(emptyConfigServerMembers)
		return nil, err
	}

	tmpConfigAddrs := c.opts.ConfigServerAddresses
	for key := range tmpConfigAddrs {
		if !strings.Contains(c.opts.ConfigServerAddresses[key], "https") && c.opts.EnableSSL {
			c.opts.ConfigServerAddresses[key] = `https://` + c.opts.ConfigServerAddresses[key]

		} else if !strings.Contains(c.opts.ConfigServerAddresses[key], "http") {
			c.opts.ConfigServerAddresses[key] = `http://` + c.opts.ConfigServerAddresses[key]
		}
	}

	err := c.Shuffle()
	if err != nil {
		openlog.Error("member shuffle is failed: " + err.Error())
		return nil, err
	}

	c.RLock()
	defer c.RUnlock()
	openlog.Debug(fmt.Sprintf("member server return %s", c.opts.ConfigServerAddresses[0]))
	return c.opts.ConfigServerAddresses, nil
}

// GetConfigs get KV from a event
func GetConfigs(actionData []byte) (map[string]interface{}, error) {
	configCenterEvent := new(Event)
	err := serializers.Decode(serializers.JSONEncoder, actionData, &configCenterEvent)
	if err != nil {
		openlog.Error(fmt.Sprintf("error in unmarshalling data on event receive with error %s", err.Error()))
		return nil, err
	}
	sourceConfig := make(map[string]interface{})
	err = serializers.Decode(serializers.JSONEncoder, []byte(configCenterEvent.Value), &sourceConfig)
	if err != nil {
		openlog.Error(fmt.Sprintf("error in unmarshalling config values %s", err.Error()))
		return nil, err
	}
	return sourceConfig, nil
}

func (c *Client) getWebSocketURL() (*url.URL, error) {
	var defaultTLS bool
	var parsedEndPoint []string
	var host string

	configCenterEntryPointList, err := c.GetConfigServer()
	if err != nil {
		openlog.Error("error in member discovery:" + err.Error())
		return nil, err
	}
	for _, server := range configCenterEntryPointList {
		parsedEndPoint = strings.Split(server, `://`)
		hostArr := strings.Split(parsedEndPoint[1], `:`)
		port := c.opts.RefreshPort
		if port == "" {
			port = "30104"
		}
		host = hostArr[0] + ":" + port
		if host == "" {
			host = "localhost"
		}
	}

	if c.wsDialer.TLSClientConfig != nil {
		defaultTLS = true
	}
	if host == "" {
		err := errors.New("host must be a URL or a host:port pair")
		openlog.Error("empty host for watch action:" + err.Error())
		return nil, err
	}
	hostURL, err := url.Parse(host)
	if err != nil || hostURL.Scheme == "" || hostURL.Host == "" {
		scheme := "ws://"
		if defaultTLS {
			scheme = "wss://"
		}
		hostURL, err = url.Parse(scheme + host)
		if err != nil {
			return nil, err
		}
		if hostURL.Path != "" && hostURL.Path != "/" {
			return nil, fmt.Errorf("host must be a URL or a host:port pair: %q", host)
		}
	}
	return hostURL, nil
}
