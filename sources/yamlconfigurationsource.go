package sources

import (
	"fmt"
	"github.com/go-fsnotify/fsnotify"
	. "github.com/servicecomb/go-archaius/core"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"reflect"
	"sync"
)

type YamlConfigurationSource struct {
	configuration map[string]interface{}
	d             *Dispatcher
	yamlFile      string
	name          string
	mu            sync.Mutex
	done          chan bool
}

func (this *YamlConfigurationSource) GetConfiguration() map[string]interface{} {
	return this.configuration
}

func (this *YamlConfigurationSource) AddDispatcher(dispatcher *Dispatcher) {
	this.d = dispatcher
}

func NewYamlConfigurationSource(name, yamlFile string) (*YamlConfigurationSource, error) {
	//Read yaml file
	yamlContent, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return nil, err
	}

	ss := yaml.MapSlice{}
	err = yaml.Unmarshal([]byte(yamlContent), &ss)
	config := retrieveItems("", ss)
	fmt.Printf("config is %s\n", config)
	source := &YamlConfigurationSource{configuration: config, yamlFile: yamlFile, name: name}

	return source, err
}

func (this *YamlConfigurationSource) GetPriority() int {
	return 12
}

func (this *YamlConfigurationSource) GetSourceName() string {
	return this.name
}

func retrieveItems(prefix string, subItems yaml.MapSlice) map[string]interface{} {
	if prefix != "" {
		prefix += "."
	}

	result := map[string]interface{}{}

	for _, item := range subItems {
		//If there are sub-items existing
		_, isSlice := item.Value.(yaml.MapSlice)
		if isSlice {
			subResult := retrieveItems(prefix+item.Key.(string), item.Value.(yaml.MapSlice))
			for k, v := range subResult {
				result[k] = v
			}
		} else {
			result[prefix+item.Key.(string)] = item.Value
			fmt.Printf("config: %s = %v\n", prefix+item.Key.(string), item.Value)
		}
	}

	return result
}

func (this *YamlConfigurationSource) AddDynamicConfigHandler(callback *ChangesCallback) error {
	if callback == nil {
		return fmt.Errorf("call back can not be nil.")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("new file watcher fail: %v", err)
		return err
	}
	err = watcher.Add(this.yamlFile)
	if err != nil {
		log.Printf("add watcher file: %s fail.", this.yamlFile)
		return err
	}
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Printf("the file %s is change for %s. reload it.", event.Name, event.Op.String())
				if event.Op != fsnotify.Write {
					log.Printf("the file change mode:  %s do not support dynamic watcher. exit and never watch again.", event.String())
					break
				}
				yamlContent, err := ioutil.ReadFile(this.yamlFile)
				if err != nil {
					log.Println(err)
					continue
				}
				ss := yaml.MapSlice{}
				err = yaml.Unmarshal([]byte(yamlContent), &ss)
				newconf := retrieveItems("", ss)
				events := this.compareUpdate(newconf, this.configuration)
				for _, e := range events {
					err := (*callback)(e)
					if err != nil {
						log.Printf("call back error: %v", err)
					}
				}
			case err := <-watcher.Errors:
				log.Println("watch file error:", err)
				break
			case this.done <- true:
				log.Println("exit wacher for %s.", this.yamlFile)
				break

			}
		}

	}()
	return nil
}

func (this *YamlConfigurationSource) compareUpdate(newconf, oldconf map[string]interface{}) []*Event {
	result := []*Event{}
	this.mu.Lock()
	defer this.mu.Unlock()
	for k, v := range oldconf {
		if _, ok := newconf[k]; !ok {
			result = append(result, &Event{EventName: k, EventType: DELETE})
			delete(this.configuration, k)
		}
		if reflect.DeepEqual(v, newconf[k]) {
			continue
		}
		result = append(result, &Event{EventName: k, EventType: UPDATE, Value: newconf[k], EventSource: k})
		this.configuration[k] = newconf[k]
	}

	for k, _ := range newconf {
		if _, ok := oldconf[k]; !ok {
			result = append(result, &Event{EventName: k, EventType: CREATE, Value: newconf[k], EventSource: k})
			this.configuration[k] = newconf[k]
		}
	}
	return result
}
