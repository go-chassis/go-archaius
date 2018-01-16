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
* Created by on 2017/7/19.
 */
// Package main for examples
package main

import (
	"fmt"
	//log "github.com/Sirupsen/logrus"
	"github.com/ServiceComb/go-archaius"
	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-archaius/sources/file-source"
	//"github.com/ServiceComb/go-archaius/sources/configcenter-source"
	//"github.com/ServiceComb/go-cc-client/member-discovery"
	"github.com/ServiceComb/go-chassis/core/lager"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//EventListener
type EventListener struct {
	Name string
}

//ConfigStruct
type ConfigStruct struct {
	Cse2 string `yaml:"cse.rest.address"`
}

//ConfigFactory
var ConfigFactory goarchaius.ConfigurationFactory

func main() {
	// init logger for archaius
	//logger = log.New()
	//logger.Formatter = new(log.JSONFormatter)
	//logger.Level = log.InfoLevel

	// create go-archaius object
	configFactory, err := goarchaius.NewConfigFactory()
	if err != nil {
		lager.Logger.Error("Error:", err)
	}
	ConfigFactory = configFactory
	// init go-archaius
	err = ConfigFactory.Init()
	if err != nil {
		lager.Logger.Error("Error:", err)
	}

	// create event receiver for configuration changes.
	// event receiver must implement core.EventListener interface
	eventListener := EventListener{Name: "eventListener1"}
	// register event receiver to go-archaius
	// regular expression support in key
	ConfigFactory.RegisterListener(eventListener, "s*")

	// get default configurations from go archaius
	// default configurations involve 1. commandline arguments 2. environment variables
	config := ConfigFactory.GetConfigurations()
	lager.Logger.Infof("======================== Default Config====================== ",
		config,
		"===========================================================\n")

	// create file source object
	fSource := filesource.NewYamlConfigurationSource()
	// add file in file source.
	// file can be regular yaml file or directory like fSource.AddFileSource("./conf", 0)
	// second argument is priority of file
	fSource.AddFileSource("./conf/name.yaml", 0)
	// add file source to go-archaius
	ConfigFactory.AddSource(fSource)

	// get default and file source configurations
	config = configFactory.GetConfigurations()
	lager.Logger.Infof("======================== Default and File source Config====================== ",
		config,
		"===========================================================\n")

	// // Steps to add config center source
	//configCenters := make([]string, 0)
	//configCenters = append(configCenters, "9.93.0.221:30103")
	//memDiscovery := memberdiscovery.NewMemDiscovery(logger.WithFields(log.Fields{
	//	"source": "member-dis",
	//}))
	//memDiscovery.Init(configCenters)
	//dimensionsInfo := "service1"
	//logger.Debug(`config client init with ` + dimensionsInfo + ` dimension info`)
	//csloger := loger.NewLogger(logger.WithFields(log.Fields{
	//	"source": "config-sebeter-logge",
	//}))
	//configCenterSource, err := configcentersource.NewConfigCenterSource(memDiscovery, dimensionsInfo, csloger)
	//if err != nil {
	//	logger.Error("invalid server uri format. ignoring config center client initialization")
	//	return
	//}
	//ConfigFactory.AddSource(configCenterSource)
	//config = ConfigFactory.GetConfigurations()
	//logger.Info("======================== Get all configurations ====================== ",
	//				config,
	//	    "===========================================================\n")

	time.Sleep(2 * time.Second)

	err = ConfigFactory.DeInit()
	if err != nil {
		lager.Logger.Error("Error:", err)
	}

	config = ConfigFactory.GetConfigurations()
	lager.Logger.Infof("======================== After Deinit Config======================\n ",
		config,
		"\n========================\n")

	//second argument is priority of file
	fSource.AddFileSource("./conf/name.yaml", 0)
	// add file source to go-`archaius
	ConfigFactory.AddSource(fSource)

	time.Sleep(1 * time.Second)
	config = ConfigFactory.GetConfigurations()
	lager.Logger.Infof("\n \n======================== after adding file source: ======================== \n", config, "======================== \n ")

	//// adding config center source
	//ConfigFactory.AddSource(configCenterSource)
	//config = ConfigFactory.GetConfigurations()
	//logger.Infoln("\n \n ======================== after adding config center source : ======================== \n", config, "======================== \n")

	// can check for key existence
	key := "name"
	if ConfigFactory.IsKeyExist(key) {
		lager.Logger.Infof(key, " key exist")
	}

	name, err := ConfigFactory.GetValue(key).ToString()
	if err != nil {
		lager.Logger.Error("get value failed with error ", err)
	} else {
		lager.Logger.Infof("Reterived value of name is ", name)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	fmt.Println("exit graceful")

}

//Event is a method get config value and logs it
func (e EventListener) Event(event *core.Event) {

	configValue := ConfigFactory.GetConfigurationByKey(event.Key)
	lager.Logger.Infof("config value ", event.Key, " | ", configValue)
}
