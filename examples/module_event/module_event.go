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
type Person struct {
	Name      string            `yaml:"name"`
	Age       int               `yaml:"age"`
	Favorites map[string]string `yaml:"favorites"`
}
type Listener struct {
	Person Person `yaml:"test.person"`
}

// Event is a method for QPS event listening
func (e *Listener) Event(events []*event.Event) {
	for i, ev := range events {
		logrus.Info(fmt.Sprintf("%dth event:%+v", i, ev))
	}
	archaius.UnmarshalConfig(e)
}

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{
		"./module_event.yaml",
	}))
	if err != nil {
		logrus.Error("Error:" + err.Error())
		return
	}
	lis := &Listener{}
	archaius.RegisterModuleListener(lis, "test.person")
	for {
		log.Printf("%+v\n", lis)
		time.Sleep(5 * time.Second)
	}
}
