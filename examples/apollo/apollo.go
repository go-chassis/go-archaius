package main

import (
	"fmt"
	agollo "github.com/Shonminh/apollo-client"
	"github.com/go-chassis/go-archaius"
	"time"
)

func main() {

	err := archaius.Init()
	fmt.Println(err)
	err = archaius.EnableApolloSource(archaius.ApolloInfo{
		ApolloAddr:        "http://127.0.0.1:8000",
		NamespaceNameList: "demo",
		AppId:             "demo-apollo",
	})
	fmt.Println(err)
	for {
		cacheMap := agollo.GetConfigCacheMap()
		for k, v:= range cacheMap {
			fmt.Printf("%v:%v\n", k, v)
		}
		time.Sleep(time.Second * 3)
	}
}
