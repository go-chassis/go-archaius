package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"time"

	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/event"
)

// Listener is a struct used for Event listener
type Listener struct {
	Key string
}

// Event is a method for QPS event listening
func (e *Listener) Event(event *event.Event) {
	logrus.Info(event.Key)
	logrus.Info(fmt.Sprintf("%v", event.Value))
	logrus.Info(event.EventType)
}

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{
		"./event.yaml",
	}))
	if err != nil {
		logrus.Error("Error:" + err.Error())
	}
	archaius.RegisterListener(&Listener{}, "age")
	for {
		log.Println(archaius.Get("age"))
		time.Sleep(5 * time.Second)
	}
}
