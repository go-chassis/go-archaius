package main

import (
	"log"
	"os"

	"github.com/arielsrv/go-archaius"
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
