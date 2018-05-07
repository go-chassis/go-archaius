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

//Package configcentersource created on 2017/6/22.
package configcentersource

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/ServiceComb/go-archaius"
	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-archaius/lager"
	"github.com/ServiceComb/go-cc-client"
	"github.com/ServiceComb/go-cc-client/serializers"
	"github.com/ServiceComb/go-chassis/config-center"
	"github.com/ServiceComb/go-chassis/core/archaius"
	"github.com/ServiceComb/go-chassis/core/common"
	"github.com/ServiceComb/go-chassis/core/config"
	"github.com/ServiceComb/go-chassis/core/endpoint-discovery"
	chassisTLS "github.com/ServiceComb/go-chassis/core/tls"
	"github.com/ServiceComb/http-client"
	"github.com/gorilla/websocket"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
	//ConfigCenterSourceConst variable of type string
	ConfigCenterSourceConst    = "ConfigCenterSource"
	configCenterSourcePriority = 0
	dimensionsInfo             = "dimensionsInfo"
	dynamicConfigAPI           = `/configuration/refresh/items`
	maxValue                   = 256
	//Name is the name of configserver
	Name = "configcenter"
)

var (
	//ConfigRefreshPath is a variable of type string
	ConfigRefreshPath = ""
)

//ConfigCenterSourceHandler handles
type ConfigCenterSourceHandler struct {
	//MemberDiscovery              memberdiscovery.MemberDiscovery
	dynamicConfigHandler         *DynamicConfigHandler
	dimensionsInfo               string
	dimensionInfoMap             map[string]string
	Configurations               map[string]interface{}
	dimensionsInfoConfiguration  map[string]map[string]interface{}
	dimensionsInfoConfigurations []map[string]map[string]interface{}
	initSuccess                  bool
	connsLock                    sync.Mutex
	sync.RWMutex
	TLSClientConfig *tls.Config
	TenantName      string
	RefreshMode     int
	RefreshInterval time.Duration
	client          *httpclient.URLClient
}

var configCenterConfig *ConfigCenterSourceHandler

//Init will initialize the Config-Center Source
func Init() {

	configCenterURL, err := isConfigCenter()
	if err != nil {
		//return nil
	}

	var enableSSL bool
	tlsConfig, tlsError := getTLSForClient(configCenterURL)
	if tlsError != nil {
		lager.Logger.Errorf(tlsError, "Get %s.%s TLS config failed, err:", Name, common.Consumer)
		//return tlsError
	}

	/*This condition added because member discovery can have multiple ip's with IsHTTPS
	having both true and false value.*/
	if tlsConfig != nil {
		enableSSL = true
	}
	refreshMode := archaius.GetInt("cse.config.client.refreshMode", common.DefaultRefreshMode)
	if refreshMode != 0 && refreshMode != 1 {
		err := errors.New(client.RefreshModeError)
		lager.Logger.Error(client.RefreshModeError, err)
		//return err
	}

	dimensionInfo := getUniqueIDForDimInfo()
	configCenterSource := NewConfigCenterSource(
		dimensionInfo, tlsConfig, config.GlobalDefinition.Cse.Config.Client.TenantName, refreshMode,
		config.GlobalDefinition.Cse.Config.Client.RefreshInterval, enableSSL)

	err = archaius.DefaultConf.ConfigFactory.AddSource(configCenterSource)
	if err != nil {
		lager.Logger.Error("failed to do add source operation!!", err)

	}

	eventHandler := EventListener{
		Name:    "EventHandler",
		Factory: archaius.DefaultConf.ConfigFactory,
	}

	//memberdiscovery.MemberDiscoveryService = memDiscovery
	archaius.DefaultConf.ConfigFactory.RegisterListener(eventHandler, "a*")

	if err := refreshGlobalConfig(); err != nil {
		lager.Logger.Error("failed to refresh global config for lb and cb", err)
		//return err
	}

	fmt.Println("Config-center Source Initialized")

}

//EventListener is a struct
type EventListener struct {
	Name    string
	Factory goarchaius.ConfigurationFactory
}

//Event is a method
func (e EventListener) Event(event *core.Event) {
	value := e.Factory.GetConfigurationByKey(event.Key)
	lager.Logger.Infof("config value %s | %s", event.Key, value)
}
func refreshGlobalConfig() error {
	err := config.ReadHystrixFromArchaius()
	if err != nil {
		return err
	}
	return config.ReadLBFromArchaius()
}

func getTLSForClient(configCenterURL string) (*tls.Config, error) {
	if !strings.Contains(configCenterURL, "://") {
		return nil, nil
	}
	ccURL, err := url.Parse(configCenterURL)
	if err != nil {
		lager.Logger.Error("Error occurred while parsing config center Server Uri", err)
		return nil, err
	}
	if ccURL.Scheme == common.HTTP {
		return nil, nil
	}

	sslTag := Name + "." + common.Consumer
	tlsConfig, sslConfig, err := chassisTLS.GetTLSConfigByService(Name, "", common.Consumer)
	if err != nil {
		if chassisTLS.IsSSLConfigNotExist(err) {
			return nil, fmt.Errorf("%s TLS mode, but no ssl config", sslTag)
		}
		return nil, err
	}
	lager.Logger.Warnf("%s TLS mode, verify peer: %t, cipher plugin: %s.",
		sslTag, sslConfig.VerifyPeer, sslConfig.CipherPlugin)

	return tlsConfig, nil
}

func isConfigCenter() (string, error) {
	configCenterURL := config.GlobalDefinition.Cse.Config.Client.ServerURI
	if configCenterURL == "" {
		ccURL, err := endpoint.GetEndpointFromServiceCenter("default", "CseConfigCenter", "latest")
		if err != nil {
			lager.Logger.Errorf(err, "empty config center endpoint, please provide the config center endpoint")
			return "", err
		}

		configCenterURL = ccURL
	}

	return configCenterURL, nil
}

func getUniqueIDForDimInfo() string {
	serviceName := config.MicroserviceDefinition.ServiceDescription.Name
	version := config.MicroserviceDefinition.ServiceDescription.Version
	appName := config.MicroserviceDefinition.AppID
	if appName == "" {
		appName = config.GlobalDefinition.AppID
	}

	if appName != "" {
		serviceName = serviceName + "@" + appName
	}

	if version != "" {
		serviceName = serviceName + "#" + version
	}

	if len(serviceName) > maxValue {
		lager.Logger.Errorf(nil, "exceeded max value %d for dimensionInfo %s with length %d", maxValue, serviceName,
			len(serviceName))
		return ""
	}

	dimeExp := `\A([^\$\%\&\+\(/)\[\]\" "\"])*\z`
	dimRegexVar, err := regexp.Compile(dimeExp)
	if err != nil {
		lager.Logger.Error("not a valid regular expression", err)
		return ""
	}

	if !dimRegexVar.Match([]byte(serviceName)) {
		lager.Logger.Errorf(nil, "invalid value for dimension info, doesnot setisfy the regular expression for dimInfo:%s",
			serviceName)
		return ""
	}

	return serviceName
}

//NewConfigCenterSource initializes all components of configuration center
func NewConfigCenterSource(dimInfo string, tlsConfig *tls.Config, tenantName string, refreshMode, refreshInterval int, enableSSL bool) core.ConfigSource {

	if configCenterConfig == nil {
		configCenterConfig = new(ConfigCenterSourceHandler)
		//configCenterConfig.MemberDiscovery = memberDiscovery
		configCenterConfig.dimensionsInfo = dimInfo
		configCenterConfig.initSuccess = true
		configCenterConfig.TLSClientConfig = tlsConfig
		configCenterConfig.TenantName = tenantName
		configCenterConfig.RefreshMode = refreshMode
		configCenterConfig.RefreshInterval = time.Second * time.Duration(refreshInterval)
		//Read the version for yaml file
		//Set Default api version to V3
		var apiVersion string
		switch config.GlobalDefinition.Cse.Config.Client.APIVersion.Version {
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
		//Update the API Base Path based on the Version
		updateAPIPath(apiVersion)

		options := &httpclient.URLClientOption{
			SSLEnabled: enableSSL,
			TLSConfig:  tlsConfig,
			Compressed: false,
			Verbose:    false,
		}
		configCenterConfig.client, _ = httpclient.GetURLClient(options)
	}
	return configCenterConfig
}

//Update the Base PATH and HEADERS Based on the version of Configcenter used.
func updateAPIPath(apiVersion string) {

	//Check for the env Name in Container to get Domain Name
	//Default value is  "default"
	projectID, isExsist := os.LookupEnv("cse.config.client.tenantName")
	if !isExsist {
		projectID = "default"
	}
	switch apiVersion {
	case "v3":
		ConfigRefreshPath = "/v3/" + projectID + dynamicConfigAPI
	case "v2":
		ConfigRefreshPath = "/configuration/v2/refresh/items"
	default:
		ConfigRefreshPath = "/v3/" + projectID + dynamicConfigAPI
	}
}

//GetConfigAPI is map
type GetConfigAPI map[string]map[string]interface{}

//CreateConfigAPI creates a configuration API
type CreateConfigAPI struct {
	DimensionInfo string                 `json:"dimensionsInfo"`
	Items         map[string]interface{} `json:"items"`
}

// ensure to implement config source
var _ core.ConfigSource = &ConfigCenterSourceHandler{}

//GetConfigurations gets a particular configuration
func (cfgSrcHandler *ConfigCenterSourceHandler) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	err := cfgSrcHandler.refreshConfigurations("")
	if err != nil {
		return nil, err
	}
	if cfgSrcHandler.RefreshMode == 1 {
		go cfgSrcHandler.refreshConfigurationsPeriodically("")
	}

	cfgSrcHandler.Lock()
	for key, value := range cfgSrcHandler.Configurations {
		configMap[key] = value
	}
	cfgSrcHandler.Unlock()
	return configMap, nil
}

//GetConfigurationsByDI gets required configurations for particular dimension info
func (cfgSrcHandler *ConfigCenterSourceHandler) GetConfigurationsByDI(dimensionInfo string) (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	err := cfgSrcHandler.refreshConfigurations(dimensionInfo)
	if err != nil {
		return nil, err
	}

	if cfgSrcHandler.RefreshMode == 1 {
		go cfgSrcHandler.refreshConfigurationsPeriodically(dimensionInfo)
	}

	cfgSrcHandler.Lock()
	for key, value := range cfgSrcHandler.dimensionsInfoConfiguration {
		configMap[key] = value
	}
	cfgSrcHandler.Unlock()
	return configMap, nil
}

func (cfgSrcHandler *ConfigCenterSourceHandler) refreshConfigurationsPeriodically(dimensionInfo string) {
	ticker := time.Tick(cfgSrcHandler.RefreshInterval)
	isConnectionFailed := false
	for range ticker {
		err := cfgSrcHandler.refreshConfigurations(dimensionInfo)
		if err == nil {
			if isConnectionFailed {
				lager.Logger.Infof("Recover configurations from config center server")
			}
			isConnectionFailed = false
		} else {
			isConnectionFailed = true
		}
	}
}

func (cfgSrcHandler *ConfigCenterSourceHandler) refreshConfigurations(dimensionInfo string) error {
	var (
		config     map[string]interface{}
		configByDI map[string]map[string]interface{}
		err        error
		events     []*core.Event
	)

	if dimensionInfo == "" {
		config, err = configcenter.DefaultClient.PullConfigs(cfgSrcHandler.dimensionsInfo, "", "", "")
		if err != nil {
			lager.Logger.Warnf("Failed to pull configurations from config center server", err) //Warn
			return err
		}
		//Populate the events based on the changed value between current config and newly received Config
		events, err = cfgSrcHandler.populateEvents(config)
	} else {
		var diInfo string
		for _, value := range cfgSrcHandler.dimensionInfoMap {
			if value == dimensionInfo {
				diInfo = dimensionInfo
			}
		}

		//configByDI, err = cfgSrcHandler.pullConfigurationsByDI(dimensionInfo)
		configByDI, err = configcenter.DefaultClient.PullConfigsByDI(dimensionInfo, diInfo)
		if err != nil {
			lager.Logger.Warnf("Failed to pull configurations from config center server", err) //Warn
			return err
		}
		//Populate the events based on the changed value between current config and newly received Config based dimension info
		events, err = cfgSrcHandler.setKeyValueByDI(configByDI, dimensionInfo)
	}

	if err != nil {
		lager.Logger.Warnf("error in generating event", err)
		return err
	}

	//Generate OnEvent Callback based on the events created
	if cfgSrcHandler.dynamicConfigHandler != nil {
		lager.Logger.Debugf("event On Receive %+v", events)
		for _, event := range events {
			cfgSrcHandler.dynamicConfigHandler.EventHandler.callback.OnEvent(event)
		}
	}

	cfgSrcHandler.Lock()
	cfgSrcHandler.updatedimensionsInfoConfigurations(dimensionInfo, configByDI, config)
	cfgSrcHandler.Unlock()

	return nil
}

func (cfgSrcHandler *ConfigCenterSourceHandler) updatedimensionsInfoConfigurations(dimensionInfo string,
	configByDI map[string]map[string]interface{}, config map[string]interface{}) {

	if dimensionInfo == "" {
		cfgSrcHandler.Configurations = config

	} else {
		if len(cfgSrcHandler.dimensionsInfoConfigurations) != 0 {
			for _, j := range cfgSrcHandler.dimensionsInfoConfigurations {
				// This condition is used to add the information of dimension info if there are 2 dimension
				if len(j) == 0 {
					cfgSrcHandler.dimensionsInfoConfigurations = append(cfgSrcHandler.dimensionsInfoConfigurations, configByDI)
				}
				for p := range j {
					if (p != dimensionInfo && len(cfgSrcHandler.dimensionInfoMap) > len(cfgSrcHandler.dimensionsInfoConfigurations)) || (len(j) == 0) {
						cfgSrcHandler.dimensionsInfoConfigurations = append(cfgSrcHandler.dimensionsInfoConfigurations, configByDI)
					}
					_, ok := j[dimensionInfo]
					if ok {
						delete(j, dimensionInfo)
						cfgSrcHandler.dimensionsInfoConfigurations = append(cfgSrcHandler.dimensionsInfoConfigurations, configByDI)
					}
				}
			}
			// This for loop to remove the emty map "map[]" from cfgSrcHandler.dimensionsInfoConfigurations
			for i, v := range cfgSrcHandler.dimensionsInfoConfigurations {
				if len(v) == 0 && len(cfgSrcHandler.dimensionsInfoConfigurations) > 1 {
					cfgSrcHandler.dimensionsInfoConfigurations = append(cfgSrcHandler.dimensionsInfoConfigurations[:i], cfgSrcHandler.dimensionsInfoConfigurations[i+1:]...)
				}
			}
		} else {
			cfgSrcHandler.dimensionsInfoConfigurations = append(cfgSrcHandler.dimensionsInfoConfigurations, configByDI)
		}

	}
}

//GetConfigurationByKey gets required configuration for a particular key
func (cfgSrcHandler *ConfigCenterSourceHandler) GetConfigurationByKey(key string) (interface{}, error) {
	cfgSrcHandler.Lock()
	configSrcVal, ok := cfgSrcHandler.Configurations[key]
	cfgSrcHandler.Unlock()
	if ok {
		return configSrcVal, nil
	}

	return nil, errors.New("key not exist")
}

//GetConfigurationByKeyAndDimensionInfo gets required configuration for a particular key and dimension pair
func (cfgSrcHandler *ConfigCenterSourceHandler) GetConfigurationByKeyAndDimensionInfo(key, dimensionInfo string) (interface{}, error) {
	var (
		configSrcVal interface{}
		actualValue  interface{}
		exist        bool
	)

	cfgSrcHandler.Lock()
	for _, v := range cfgSrcHandler.dimensionsInfoConfigurations {
		value, ok := v[dimensionInfo]
		if ok {
			actualValue, exist = value[key]
		}
	}
	cfgSrcHandler.Unlock()

	if exist {
		configSrcVal = actualValue
		return configSrcVal, nil
	}

	return nil, errors.New("key not exist")
}

//AddDimensionInfo adds dimension info for a configuration
func (cfgSrcHandler *ConfigCenterSourceHandler) AddDimensionInfo(dimensionInfo string) (map[string]string, error) {
	if len(cfgSrcHandler.dimensionInfoMap) == 0 {
		cfgSrcHandler.dimensionInfoMap = make(map[string]string)
	}

	for i := range cfgSrcHandler.dimensionInfoMap {
		if i == dimensionInfo {
			lager.Logger.Errorf(nil, "dimension info allready exist")
			return cfgSrcHandler.dimensionInfoMap, errors.New("dimension info allready exist")
		}
	}

	cfgSrcHandler.dimensionInfoMap[dimensionInfo] = dimensionInfo

	return cfgSrcHandler.dimensionInfoMap, nil
}

//GetSourceName returns name of the configuration
func (*ConfigCenterSourceHandler) GetSourceName() string {
	return ConfigCenterSourceConst
}

//GetPriority returns priority of a configuration
func (*ConfigCenterSourceHandler) GetPriority() int {
	return configCenterSourcePriority
}

//DynamicConfigHandler dynamically handles a configuration
func (cfgSrcHandler *ConfigCenterSourceHandler) DynamicConfigHandler(callback core.DynamicConfigCallback) error {
	if cfgSrcHandler.initSuccess != true {
		return errors.New("config center source initialization failed")
	}

	dynCfgHandler, err := newDynConfigHandlerSource(cfgSrcHandler, callback)
	if err != nil {
		lager.Logger.Error("failed to initialize dynamic config center ConfigCenterSourceHandler", err)
		return errors.New("failed to initialize dynamic config center ConfigCenterSourceHandler")
	}
	cfgSrcHandler.dynamicConfigHandler = dynCfgHandler

	if cfgSrcHandler.RefreshMode == 0 {
		// Pull All the configuration for the first time.
		cfgSrcHandler.refreshConfigurations("")
		//Start a web socket connection to recieve change events.
		dynCfgHandler.startDynamicConfigHandler()
	}

	return nil
}

//Cleanup cleans the particular configuration up
func (cfgSrcHandler *ConfigCenterSourceHandler) Cleanup() error {
	cfgSrcHandler.connsLock.Lock()
	defer cfgSrcHandler.connsLock.Unlock()

	if cfgSrcHandler.dynamicConfigHandler != nil {
		cfgSrcHandler.dynamicConfigHandler.Cleanup()
	}

	cfgSrcHandler.dynamicConfigHandler = nil
	cfgSrcHandler.Configurations = nil

	return nil
}

func (cfgSrcHandler *ConfigCenterSourceHandler) populateEvents(updatedConfig map[string]interface{}) ([]*core.Event, error) {
	events := make([]*core.Event, 0)
	newConfig := make(map[string]interface{})
	cfgSrcHandler.Lock()
	defer cfgSrcHandler.Unlock()

	currentConfig := cfgSrcHandler.Configurations

	// generate create and update event
	for key, value := range updatedConfig {
		newConfig[key] = value
		currentValue, ok := currentConfig[key]
		if !ok { // if new configuration introduced
			events = append(events, cfgSrcHandler.constructEvent(core.Create, key, value))
		} else if !reflect.DeepEqual(currentValue, value) {
			events = append(events, cfgSrcHandler.constructEvent(core.Update, key, value))
		}
	}

	// generate delete event
	for key, value := range currentConfig {
		_, ok := newConfig[key]
		if !ok { // when old config not present in new config
			events = append(events, cfgSrcHandler.constructEvent(core.Delete, key, value))
		}
	}

	// update with latest config
	cfgSrcHandler.Configurations = newConfig

	return events, nil
}

func (cfgSrcHandler *ConfigCenterSourceHandler) setKeyValueByDI(updatedConfig map[string]map[string]interface{}, dimensionInfo string) ([]*core.Event, error) {
	events := make([]*core.Event, 0)
	newConfigForDI := make(map[string]map[string]interface{})
	cfgSrcHandler.Lock()
	defer cfgSrcHandler.Unlock()

	currentConfig := cfgSrcHandler.dimensionsInfoConfiguration

	// generate create and update event
	for key, value := range updatedConfig {
		if key == dimensionInfo {
			newConfigForDI[key] = value
			for k, v := range value {
				if len(currentConfig) == 0 {
					events = append(events, cfgSrcHandler.constructEvent(core.Create, k, v))
				}
				for diKey, val := range currentConfig {
					if diKey == dimensionInfo {
						currentValue, ok := val[k]
						if !ok { // if new configuration introduced
							events = append(events, cfgSrcHandler.constructEvent(core.Create, k, v))
						} else if currentValue != v {
							events = append(events, cfgSrcHandler.constructEvent(core.Update, k, v))
						}
					}
				}
			}
		}
	}

	// generate delete event
	for key, value := range currentConfig {
		if key == dimensionInfo {
			for k, v := range value {
				for _, val := range newConfigForDI {
					_, ok := val[k]
					if !ok {
						events = append(events, cfgSrcHandler.constructEvent(core.Delete, k, v))
					}
				}
			}
		}
	}

	// update with latest config
	cfgSrcHandler.dimensionsInfoConfiguration = newConfigForDI

	return events, nil
}

func (cfgSrcHandler *ConfigCenterSourceHandler) constructEvent(eventType string, key string, value interface{}) *core.Event {
	newEvent := new(core.Event)
	newEvent.EventSource = ConfigCenterSourceConst
	newEvent.EventType = eventType
	newEvent.Key = key
	newEvent.Value = value

	return newEvent
}

//DynamicConfigHandler is a struct
type DynamicConfigHandler struct {
	dimensionsInfo string
	EventHandler   *ConfigCenterEventHandler
	dynamicLock    sync.Mutex
	wsDialer       *websocket.Dialer
	wsConnection   *websocket.Conn
	//memberDiscovery memberdiscovery.MemberDiscovery
}

func newDynConfigHandlerSource(cfgSrc *ConfigCenterSourceHandler, callback core.DynamicConfigCallback) (*DynamicConfigHandler, error) {
	eventHandler := newConfigCenterEventHandler(cfgSrc, callback)
	dynCfgHandler := new(DynamicConfigHandler)
	dynCfgHandler.EventHandler = eventHandler
	dynCfgHandler.dimensionsInfo = cfgSrc.dimensionsInfo
	dynCfgHandler.wsDialer = &websocket.Dialer{
		TLSClientConfig:  cfgSrc.TLSClientConfig,
		HandshakeTimeout: defaultTimeout,
	}
	//dynCfgHandler.memberDiscovery = cfgSrc.MemberDiscovery
	return dynCfgHandler, nil
}

func (dynHandler *DynamicConfigHandler) getWebSocketURL() (*url.URL, error) {

	var defaultTLS bool
	var parsedEndPoint []string
	var host string

	configCenterEntryPointList, err := configcenter.DefaultClient.GetConfigServer()
	if err != nil {
		lager.Logger.Error("error in member discovery", err)
		return nil, err
	}
	activeEndPointList, err := configcenter.DefaultClient.GetWorkingConfigCenterIP(configCenterEntryPointList)
	if err != nil {
		lager.Logger.Error("failed to get ip list", err)
	}
	for _, server := range activeEndPointList {
		parsedEndPoint = strings.Split(server, `://`)
		hostArr := strings.Split(parsedEndPoint[1], `:`)
		port := config.GlobalDefinition.Cse.Config.Client.RefreshPort
		if port == "" {
			port = "30104"
		}
		host = hostArr[0] + ":" + port
		if host == "" {
			host = "localhost"
		}
	}

	if dynHandler.wsDialer.TLSClientConfig != nil {
		defaultTLS = true
	}
	if host == "" {
		err := errors.New("host must be a URL or a host:port pair")
		lager.Logger.Error("empty host for watch action", err)
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

func (dynHandler *DynamicConfigHandler) startDynamicConfigHandler() error {
	parsedDimensionInfo := strings.Replace(dynHandler.dimensionsInfo, "#", "%23", -1)
	refreshConfigPath := ConfigRefreshPath + `?` + dimensionsInfo + `=` + parsedDimensionInfo
	if dynHandler != nil && dynHandler.wsDialer != nil {
		/*-----------------
		1. Decide on the URL
		2. Create WebSocket Connection
		3. Call KeepAlive in seperate thread
		3. Generate events on Recieve Data
		*/
		baseURL, err := dynHandler.getWebSocketURL()
		if err != nil {
			error := errors.New("error in getting default server info")
			return error
		}
		url := baseURL.String() + refreshConfigPath
		dynHandler.dynamicLock.Lock()
		dynHandler.wsConnection, _, err = dynHandler.wsDialer.Dial(url, nil)
		if err != nil {
			dynHandler.dynamicLock.Unlock()
			return fmt.Errorf("watching config-center dial catch an exception error:%s", err.Error())
		}
		dynHandler.dynamicLock.Unlock()
		keepAlive(dynHandler.wsConnection, 15*time.Second)
		go func() error {
			for {
				messageType, message, err := dynHandler.wsConnection.ReadMessage()
				if err != nil {
					break
				}
				if messageType == websocket.TextMessage {
					dynHandler.EventHandler.OnReceive(message)
				}
			}
			err = dynHandler.wsConnection.Close()
			if err != nil {
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

//Cleanup cleans particular dynamic configuration ConfigCenterSourceHandler up
func (dynHandler *DynamicConfigHandler) Cleanup() error {
	dynHandler.dynamicLock.Lock()
	defer dynHandler.dynamicLock.Unlock()
	if dynHandler.wsConnection != nil {
		dynHandler.wsConnection.Close()
	}
	dynHandler.wsConnection = nil
	return nil
}

//ConfigCenterEventHandler handles a event of a configuration center
type ConfigCenterEventHandler struct {
	configSource *ConfigCenterSourceHandler
	callback     core.DynamicConfigCallback
}

//ConfigCenterEvent stores info about an configuration center event
type ConfigCenterEvent struct {
	Action string `json:"action"`
	Value  string `json:"value"`
}

func newConfigCenterEventHandler(cfgSrc *ConfigCenterSourceHandler, callback core.DynamicConfigCallback) *ConfigCenterEventHandler {
	eventHandler := new(ConfigCenterEventHandler)
	eventHandler.configSource = cfgSrc
	eventHandler.callback = callback
	return eventHandler
}

//OnConnect is a method
func (*ConfigCenterEventHandler) OnConnect() {
	return
}

//OnConnectionClose is a method
func (*ConfigCenterEventHandler) OnConnectionClose() {
	return
}

//OnReceive initializes all necessary components for a configuration center
func (eventHandler *ConfigCenterEventHandler) OnReceive(actionData []byte) {
	configCenterEvent := new(ConfigCenterEvent)
	err := serializers.Decode(serializers.JsonEncoder, actionData, &configCenterEvent)
	if err != nil {
		lager.Logger.Errorf(err, fmt.Sprintf("error in unmarshalling data on event receive with error %s", err.Error()))
		return
	}

	sourceConfig := make(map[string]interface{})
	err = serializers.Decode(serializers.JsonEncoder, []byte(configCenterEvent.Value), &sourceConfig)
	if err != nil {
		lager.Logger.Errorf(err, fmt.Sprintf("error in unmarshalling config values %s", err.Error()))
		return
	}

	events, err := eventHandler.configSource.populateEvents(sourceConfig)
	if err != nil {
		lager.Logger.Error("error in generating event", err)
		return
	}

	lager.Logger.Debugf("event On Receive", events)
	for _, event := range events {
		eventHandler.callback.OnEvent(event)
	}

	return
}
