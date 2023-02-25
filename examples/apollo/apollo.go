package main

import (
	"fmt"
	"time"

	agollo "github.com/Shonminh/apollo-client"
	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/event"
	"github.com/arielsrv/go-archaius/source/apollo"
	_ "github.com/arielsrv/go-archaius/source/apollo"
	"github.com/go-chassis/openlog"
)

type Listener struct {
	Key string
}

func (li *Listener) Event(event *event.Event) {
	fmt.Printf("listen:%+v", *event)
	openlog.Info(event.Key)
	openlog.Info(fmt.Sprintf("%v\n", event.Value))
	openlog.Info(event.EventType)
}

func main() {

	err := archaius.Init(archaius.WithRemoteSource(archaius.ApolloSource, &archaius.RemoteInfo{
		URL: "http://127.0.0.1:8000",
		DefaultDimension: map[string]string{
			apollo.AppID:         "demo-apollo",
			apollo.NamespaceList: "demo",
		},
	}))
	fmt.Println(err)
	// register listener, key is different from which in apollo web page, it's format is like {namespace}.{apollo_key}
	err = archaius.RegisterListener(&Listener{}, "demo.xxx")
	fmt.Println(err)
	for {
		cacheMap := agollo.GetConfigCacheMap()
		for k, v := range cacheMap {
			fmt.Printf("%v:%v\n", k, v)
		}
		time.Sleep(time.Second * 3)
	}
}
