package main

import (
	"github.com/go-chassis/go-archaius"
	"github.com/go-mesh/openlogging"
	"log"
)

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{"f1.yaml"}))
	if err != nil {
		openlogging.GetLogger().Error("Error:" + err.Error())
	}
	log.Println(archaius.Get("age"))
	log.Println(archaius.Get("name"))
	err = archaius.AddFile("f2.yaml")
	if err != nil {
		log.Panicln(err)
	}
	log.Println(archaius.Get("age"))
	log.Println(archaius.Get("name"))

}
