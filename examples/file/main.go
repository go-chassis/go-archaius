package main

import (
	"github.com/sirupsen/logrus"
	"log"

	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/source/util"
)

func main() {
	err := archaius.Init(archaius.WithRequiredFiles([]string{"./dir", "f1.yaml"}))
	if err != nil {
		logrus.Error("Error:" + err.Error())
	}
	log.Println(archaius.Get("age"))
	log.Println(archaius.Get("name"))
	log.Println(archaius.Get("favorite"))
	log.Println(archaius.Get("c"))
	log.Println(archaius.Get("b"))
	err = archaius.AddFile("f2.yaml")
	if err != nil {
		log.Panicln(err)
	}
	log.Println(archaius.Get("age"))
	log.Println(archaius.Get("name"))

	err = archaius.AddFile("f3.yaml", archaius.WithFileHandler(util.FileHandler(util.UseFileNameAsKeyContentAsValue)))
	if err != nil {
		log.Panicln(err)
	}
	log.Println(archaius.GetString("f3.yaml", ""))
}
