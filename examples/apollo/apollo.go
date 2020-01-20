package main

import (
	"fmt"
	agollo "github.com/Shonminh/apollo-client"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-mesh/openlogging"
	"time"
)

type Listener struct {
	Key string
}

func (li *Listener) Event(event *event.Event)  {
	fmt.Printf("listen:%+v", *event)
	openlogging.GetLogger().Info(event.Key)
	openlogging.GetLogger().Infof(fmt.Sprintf("%v", event.Value))
	openlogging.GetLogger().Info(event.EventType)
}

func main() {

	err := archaius.Init()
	fmt.Println(err)
	// register listener, key is different from which in apollo web page, it's format is like {namespace}.{apollo_key}
	err = archaius.RegisterListener(&Listener{}, "demo.xxx")

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
