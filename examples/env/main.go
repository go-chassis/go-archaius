package main

import (
	"github.com/go-chassis/go-archaius"
	"log"
	"os"
)

func main() {
	os.Setenv("cred_db_user", "root")
	os.Setenv("cred_db_pwd", "root")
	err := archaius.Init(archaius.WithRequiredFiles([]string{"f1.yaml"}),
		archaius.WithENVSource())
	if err != nil {
		panic(err)
	}
	log.Println(archaius.Get("cred.db.user"))
	log.Println(archaius.Get("cred.db.pwd"))

}
