package apollo

import (
	apollo "github.com/Shonminh/apollo-client"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/go-archaius/source"
	"github.com/pkg/errors"
	"sync"
)

// Source apollo source
type Source struct {
	priority int
	sync.RWMutex
	eventHandler source.EventHandler
}

const (
	defaultApolloSourcePriority = 0 // default priority is 0
	apolloSourceName            = "ApolloConfigSource"
)

var (
	gStartApolloOnce sync.Once
)

type NamespaceParser func(originalKey string) (namespaceName string)

// NewApolloSource get a apollo source singleton, and pull configs at once after init apollo client.
func NewApolloSource(opts ...apollo.Option) (source.ConfigSource, error) {
	as := new(Source)
	as.priority = defaultApolloSourcePriority
	if err := apollo.Init(opts...); err != nil {
		return nil, errors.WithMessage(err, "apollo client init failed")
	}
	return as, nil
}

// GetConfigurations get config cache map from apollo client.
func (as *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{}) // 该config的value表示的source
	as.Lock()
	apolloCache := apollo.GetConfigCacheMap()
	for k := range apolloCache {
		configMap[k] = apolloSourceName
	}
	as.Unlock()
	return configMap, nil
}

// GetConfigurationByKey get config by key, key's format is: {namespace}.field1.field2
func (as *Source) GetConfigurationByKey(key string) (interface{}, error) {
	value, err := apollo.GetConfigByKey(key)
	if err != nil {
		return nil, errors.WithMessage(err, "GetConfigByKey")
	}
	return value, nil
}

// Watch register change event handler and start refresh configs interval.
func (as *Source) Watch(callBack source.EventHandler) error {
	as.eventHandler = callBack
	apollo.RegChangeEventHandler(as.UpdateCallback)
	// start refresh routine once
	gStartApolloOnce.Do(func() {
		go apollo.Start()
	})
	return nil
}

// GetPriority get priority
func (as *Source) GetPriority() int {
	return as.priority
}

// SetPriority set priority
func (as *Source) SetPriority(priority int) {
	as.priority = priority
}

// Cleanup clean apollo cache from apollo client
func (as *Source) Cleanup() error {
	apollo.Cleanup()
	return nil
}

// GetSourceName get source name
func (as *Source) GetSourceName() string {
	return apolloSourceName
}

// no use
func (as *Source) AddDimensionInfo(labels map[string]string) error {
	return nil
}

// no use
func (as *Source) Set(key string, value interface{}) error {
	return nil
}

// no use
func (as *Source) Delete(key string) error {
	return nil
}

// UpdateCallback callback function when config updates
func (as *Source) UpdateCallback(apolloEvent *apollo.ChangeEvent) error {
	if as.eventHandler != nil {
		for _, c := range apolloEvent.Changes {
			eventType := transformEventType(c.ChangeType)
			if eventType == "" {
				continue
			}

			e := &event.Event{
				EventSource: apolloSourceName,
				EventType:   eventType,
				Key:         apolloEvent.Namespace + "." + c.Key, // to make sure key is prefix with namespace
				Value:       c.NewValue,
			}
			as.eventHandler.OnEvent(e)
		}
	}
	return nil
}

// transformEventType transform change type
func transformEventType(changeType apollo.ConfigChangeType) string {
	switch changeType {
	case apollo.ADDED:
		return event.Create
	case apollo.MODIFIED:
		return event.Update
	case apollo.DELETED:
		return event.Delete
	}
	return ""
}
