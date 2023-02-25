package main

import (
	"fmt"
	"log"
	"time"

	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/event"
	"github.com/go-chassis/openlog"
)

// Listener is a struct used for Event listener
type Listener struct {
	Key string
}

// Event is a method for QPS event listening
func (e *Listener) Event(event *event.Event) {
	openlog.Info(event.Key)
	openlog.Info(fmt.Sprintf("%v", event.Value))
	openlog.Info(event.EventType)
}

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{
		"./event.yaml",
	}))
	if err != nil {
		openlog.Error("Error:" + err.Error())
	}
	archaius.RegisterListener(&Listener{}, "age")
	for {
		log.Println(archaius.Get("age"))
		time.Sleep(5 * time.Second)
	}
}
