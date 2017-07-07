package main

import (
	"fmt"
	"github.com/servicecomb/go-archaius"
	"github.com/servicecomb/go-archaius/core"
	"github.com/servicecomb/go-archaius/sources"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	//init if you do not want to use the default env&yaml source.
	core.DefaultSources = []core.ConfigurationSource{}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	confFile := filepath.Join(pwd, "conf/name.yaml")
	yamlsource, err := sources.NewYamlConfigurationSource("nameconfig", confFile)
	if err != nil {
		panic(err)
	}
	core.DefaultSources = append(core.DefaultSources, yamlsource)
	factory := goarchaius.NewConfigurationFactory()
	factory.Init()
	callback := func(e *core.Event) {
		log.Printf("the value of %s is change to %v", e.EventName, e.Value)
	}
	factory.RegisterListener("spouse.age", (*core.EventCallback)(&callback))
	age := factory.GetConfigurationByKey("spouse.age")
	log.Printf("get the detail key: spouse.age, value is: %v", age)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	fmt.Println("exit graceful.")
}
