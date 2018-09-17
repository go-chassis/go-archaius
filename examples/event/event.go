package main

import (
	"fmt"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/core"
	"github.com/go-chassis/go-archaius/sources/file-source"
	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/go-mesh/openlogging"
	"log"
	"time"
)

//Listener is a struct used for Event listener
type Listener struct {
	Key string
}

//Event is a method for QPS event listening
func (e *Listener) Event(event *core.Event) {
	openlogging.GetLogger().Info(event.Key)
	openlogging.GetLogger().Infof(fmt.Sprintf("%s", event.Value))
	openlogging.GetLogger().Info(event.EventType)
}

//
func main() {
	lager.Initialize("", "DEBUG", "", "size", true, 1, 10, 7)
	configFactory, err := goarchaius.NewConfigFactory(nil)
	if err != nil {
		openlogging.GetLogger().Error("Error:" + err.Error())
	}
	// init go-archaius
	err = configFactory.Init()
	if err != nil {
		openlogging.GetLogger().Error("Error:" + err.Error())
	}

	fSource := filesource.NewYamlConfigurationSource()
	// add file in file source.
	// file can be regular yaml file or directory like fSource.AddFileSource("./conf", 0)
	// second argument is priority of file
	fSource.AddFileSource("./event.yaml", 0)
	// add file source to go-archaius
	configFactory.AddSource(fSource)
	configFactory.RegisterListener(&Listener{}, "age")
	for {
		log.Println(configFactory.GetConfigurationByKey("age"))
		time.Sleep(20 * time.Second)
	}
}
