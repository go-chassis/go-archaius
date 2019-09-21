package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-mesh/openlogging"
)

//Listener is a struct used for Event listener
type Listener struct {
	Key string
}

//Event is a method for QPS event listening
func (e *Listener) Event(event *event.Event) {
	openlogging.GetLogger().Info(event.Key)
	openlogging.GetLogger().Infof(fmt.Sprintf("%s", event.Value))
	openlogging.GetLogger().Info(event.EventType)
}

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{
		"./event.yaml",
	}))
	if err != nil {
		openlogging.GetLogger().Error("Error:" + err.Error())
	}
	archaius.RegisterListener(&Listener{}, "age")
	for {
		log.Println(archaius.Get("age"))
		time.Sleep(5 * time.Second)
	}
}
