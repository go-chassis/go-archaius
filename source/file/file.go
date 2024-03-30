/*
 * Copyright 2017 Huawei Technologies Co., Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//Package filesource created on 2017/6/22.
package filesource

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/go-archaius/source"
	"github.com/go-chassis/go-archaius/source/util"
	"github.com/go-chassis/openlog"
)

const (
	//FileConfigSourceConst is a variable of type string
	FileConfigSourceConst = "FileSource"
	fileSourcePriority    = 4
	//DefaultFilePriority is a variable of type string
	DefaultFilePriority = 0
)

//FileSourceTypes is a string
type FileSourceTypes string

const (
	//RegularFile is a variable of type string
	RegularFile FileSourceTypes = "RegularFile"
	//Directory is a variable of type string
	Directory FileSourceTypes = "Directory"
	//InvalidFileType is a variable of type string
	InvalidFileType FileSourceTypes = "InvalidType"
)

//ConfigInfo is s struct
type ConfigInfo struct {
	FilePath string
	Value    interface{}
}

//Source is file source
type Source struct {
	Configurations map[string]*ConfigInfo
	files          []file
	fileHandlers   map[string]util.FileHandler
	watchPool      *watch
	filelock       sync.Mutex
	priority       int
	sync.RWMutex
}

type file struct {
	filePath string
	priority uint32
}

type watch struct {
	//files   []string
	watcher    *fsnotify.Watcher
	callback   source.EventHandler
	fileSource *Source
	sync.RWMutex
}

/*
	accepts files and directories as file-source
  		1. Directory: all files considered as file source
  		2. File: specified file considered as file source

  	TODO: Currently file sources priority not considered. if key conflicts then latest key will get considered
*/

//FileSource is a interface
type FileSource interface {
	source.ConfigSource
	AddFile(filePath string, priority uint32, handler util.FileHandler) error
}

//NewFileSource creates a source which can handler local files
func NewFileSource() FileSource {
	fileConfigSource := new(Source)
	fileConfigSource.priority = fileSourcePriority
	fileConfigSource.files = make([]file, 0)
	fileConfigSource.fileHandlers = make(map[string]util.FileHandler)
	return fileConfigSource
}

//AddFile add file and manage configs
func (fSource *Source) AddFile(p string, priority uint32, handle util.FileHandler) error {
	path, err := filepath.Abs(p)
	if err != nil {
		return err
	}
	// check existence of file
	fs, err := os.Open(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("[%s] file not exist", path)
	}
	defer fs.Close()

	// prevent duplicate file source
	if fSource.isFileSrcExist(path) {
		return nil
	}
	fSource.fileHandlers[path] = handle
	fileType := fileType(fs)
	switch fileType {
	case Directory:
		// handle Directory input. Include all files as file source.
		err := fSource.handleDirectory(fs, priority, handle)
		if err != nil {
			openlog.Error(fmt.Sprintf("Failed to handle directory [%s] %s", path, err))
			return err
		}
	case RegularFile:
		// handle file and include as file source.
		err := fSource.handleFile(fs, priority, handle)
		if err != nil {
			openlog.Error(fmt.Sprintf("Failed to handle file [%s] [%s]", path, err))
			return err
		}
	case InvalidFileType:
		openlog.Error(fmt.Sprintf("File type of [%s] not supported: %s", path, err))
		return fmt.Errorf("file type of [%s] not supported", path)
	}

	if fSource.watchPool != nil {
		fSource.watchPool.AddWatchFile(path)
	}

	return nil
}

func (fSource *Source) isFileSrcExist(filePath string) bool {
	var exist bool
	for _, file := range fSource.files {
		if filePath == file.filePath {
			return true
		}
	}
	return exist
}

func fileType(fs *os.File) FileSourceTypes {
	fileInfo, err := fs.Stat()
	if err != nil {
		return InvalidFileType
	}

	fileMode := fileInfo.Mode()

	if fileMode.IsDir() {
		return Directory
	} else if fileMode.IsRegular() {
		return RegularFile
	}

	return InvalidFileType
}

func (fSource *Source) handleDirectory(dir *os.File, priority uint32, handle util.FileHandler) error {

	filesInfo, err := dir.Readdir(-1)
	if err != nil {
		return errors.New("failed to read Directory contents")
	}

	for _, fileInfo := range filesInfo {
		filePath := filepath.Join(dir.Name(), fileInfo.Name())

		fs, err := os.Open(filePath)
		if err != nil {
			openlog.Error(fmt.Sprintf("error in file open for %s file", err.Error()))
			continue
		}

		err = fSource.handleFile(fs, priority, handle)
		if err != nil {
			openlog.Error(fmt.Sprintf("error processing %s file source handler with error : %s ", fs.Name(),
				err.Error()))
		}
		fs.Close()

	}

	return nil
}

func (fSource *Source) handleFile(file *os.File, priority uint32, handle util.FileHandler) error {
	Content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return err
	}
	var config map[string]interface{}
	if handle != nil {
		config, err = handle(file.Name(), Content)
	} else {
		config, err = util.Convert2JavaProps(file.Name(), Content)
	}
	if err != nil {
		return fmt.Errorf("failed to pull configurations from [%s] file, %s", file.Name(), err)
	}

	err = fSource.handlePriority(file.Name(), priority)
	if err != nil {
		return fmt.Errorf("failed to handle priority of [%s], %s", file.Name(), err)
	}

	events := fSource.compareUpdate(config, file.Name())
	if fSource.watchPool != nil && fSource.watchPool.callback != nil { // if file source already added and try to add
		for _, e := range events {
			fSource.watchPool.callback.OnEvent(e)
		}
		fSource.watchPool.callback.OnModuleEvent(events)
	}

	return nil
}

func (fSource *Source) handlePriority(filePath string, priority uint32) error {
	fSource.Lock()
	newFilePriority := make([]file, 0)
	var prioritySet bool
	for _, f := range fSource.files {

		if f.filePath == filePath && f.priority == priority {
			prioritySet = true
			newFilePriority = append(newFilePriority, file{
				filePath: filePath,
				priority: priority,
			})
		}
		newFilePriority = append(newFilePriority, f)
	}

	if !prioritySet {
		newFilePriority = append(newFilePriority, file{
			filePath: filePath,
			priority: priority,
		})
	}

	fSource.files = newFilePriority
	fSource.Unlock()

	return nil
}

//GetConfigurations get all configs
func (fSource *Source) GetConfigurations() (map[string]interface{}, error) {
	configMap := make(map[string]interface{})

	fSource.Lock()
	defer fSource.Unlock()
	for key, confInfo := range fSource.Configurations {
		if confInfo == nil {
			configMap[key] = nil
			continue
		}

		configMap[key] = confInfo.Value
	}

	return configMap, nil
}

//GetConfigurationByKey get one key value
func (fSource *Source) GetConfigurationByKey(key string) (interface{}, error) {
	fSource.RLock()
	defer fSource.RUnlock()

	confInfo, ok := fSource.Configurations[key]
	if !ok || confInfo == nil {
		return nil, source.ErrKeyNotExist
	}

	return confInfo.Value, nil
}

//GetSourceName get name of source
func (*Source) GetSourceName() string {
	return FileConfigSourceConst
}

//GetPriority get precedence
func (fSource *Source) GetPriority() int {
	return fSource.priority
}

//SetPriority custom priority
func (fSource *Source) SetPriority(priority int) {
	fSource.priority = priority
}

//Watch watch change event
func (fSource *Source) Watch(callback source.EventHandler) error {
	if callback == nil {
		return errors.New("call back can not be nil")
	}

	watchPool, err := newWatchPool(callback, fSource)
	if err != nil {
		return err
	}

	fSource.watchPool = watchPool

	go fSource.watchPool.startWatchPool()

	return nil
}

func newWatchPool(callback source.EventHandler, cfgSrc *Source) (*watch, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		openlog.Error("New file watcher failed:" + err.Error())
		return nil, err
	}

	watch := new(watch)
	watch.callback = callback
	watch.fileSource = cfgSrc
	watch.watcher = watcher
	openlog.Info("create new watcher")
	return watch, nil
}

func (wth *watch) startWatchPool() {
	go wth.watchFile()
	for _, file := range wth.fileSource.files {
		f, err := filepath.Abs(file.filePath)
		if err != nil {
			openlog.Error(fmt.Sprintf("failed to get Directory info from: %s file: %s", file.filePath, err))
			return
		}

		err = wth.watcher.Add(f)
		if err != nil {
			openlog.Error(fmt.Sprintf("add watcher file: %+v fail %s", file, err))
			return
		}
	}
}

func (wth *watch) AddWatchFile(filePath string) {
	err := wth.watcher.Add(filePath)
	if err != nil {
		openlog.Error(fmt.Sprintf("add watcher file: %s fail: %s", filePath, err))
		return
	}
}

func (wth *watch) watchFile() {
	for {
		select {
		case event, ok := <-wth.watcher.Events:
			if !ok {
				openlog.Warn("file watcher stop")
				return
			}

			if strings.HasSuffix(event.Name, ".swx") || strings.HasSuffix(event.Name, ".swp") || strings.HasSuffix(event.Name, "~") {
				//ignore
				continue
			}
			openlog.Debug(fmt.Sprintf("file event %s, operation is %d. reload it.", event.Name, event.Op))

			if event.Op == fsnotify.Remove {
				openlog.Warn(fmt.Sprintf("the file change mode: %s, continue", event.String()))
				continue
			}

			if event.Op == fsnotify.Rename {
				openlog.Debug("file renamed")
				wth.watcher.Remove(event.Name)
				// check existence of file
				_, err := os.Open(event.Name)
				if os.IsNotExist(err) {
					openlog.Warn(fmt.Sprintf("[%s] file does not exist so not able to watch further:%s", event.Name, err))
				} else {
					wth.AddWatchFile(event.Name)
				}

				continue
			}

			if event.Op == fsnotify.Create {
				openlog.Debug("file created")
				time.Sleep(time.Millisecond)
			}
			handle := wth.fileSource.fileHandlers[event.Name]
			if handle == nil {
				openlog.Debug("user default file handler")
				handle = util.Convert2JavaProps
			}
			content, err := ioutil.ReadFile(event.Name)
			if err != nil {
				openlog.Error("read file error " + err.Error())
				continue
			}

			newConf, err := handle(event.Name, content)
			if err != nil {
				openlog.Error("convert error " + err.Error())
				continue
			}
			openlog.Debug(fmt.Sprintf("new config: %v", newConf))
			events := wth.fileSource.compareUpdate(newConf, event.Name)
			openlog.Debug(fmt.Sprintf("generated events %v", events))
			if len(events) > 0 { //avoid OnModuleEvent empty events error
				for _, e := range events {
					wth.callback.OnEvent(e)
				}
				wth.callback.OnModuleEvent(events)
			}

		case err := <-wth.watcher.Errors:
			openlog.Debug(fmt.Sprintf("watch file error: %s", err))
			return
		}
	}

}

func (fSource *Source) compareUpdate(configs map[string]interface{}, filePath string) []*event.Event {
	events := make([]*event.Event, 0)
	fileConfs := make(map[string]*ConfigInfo)
	if fSource == nil {
		return nil
	}

	fSource.Lock()
	defer fSource.Unlock()

	var filePathPriority uint32 = math.MaxUint32
	for _, file := range fSource.files {
		if file.filePath == filePath {
			filePathPriority = file.priority
		}
	}

	if filePathPriority == math.MaxUint32 {
		return nil
	}

	// update and delete with latest configs

	for key, confInfo := range fSource.Configurations {
		if confInfo == nil {
			continue
		}

		switch confInfo.FilePath {
		case filePath:
			newConfValue, ok := configs[key]
			if !ok {
				events = append(events, &event.Event{EventSource: FileConfigSourceConst, Key: key,
					EventType: event.Delete, Value: confInfo.Value})
				continue
			} else if reflect.DeepEqual(confInfo.Value, newConfValue) {
				fileConfs[key] = confInfo
				continue
			}

			confInfo.Value = newConfValue
			fileConfs[key] = confInfo

			events = append(events, &event.Event{EventSource: FileConfigSourceConst, Key: key,
				EventType: event.Update, Value: newConfValue})

		default: // configuration file not same
			// no need to handle delete scenario
			// only handle if configuration conflicts between two sources
			newConfValue, ok := configs[key]
			if ok {
				var priority uint32 = math.MaxUint32
				for _, file := range fSource.files {
					if file.filePath == confInfo.FilePath {
						priority = file.priority
					}
				}

				if priority == filePathPriority {
					fileConfs[key] = confInfo
					openlog.Info(fmt.Sprintf("Two files have same priority. keeping %s value", confInfo.FilePath))

				} else if filePathPriority < priority { // lower the vale higher is the priority
					confInfo.Value = newConfValue
					fileConfs[key] = confInfo
					events = append(events, &event.Event{EventSource: FileConfigSourceConst,
						Key: key, EventType: event.Update, Value: newConfValue})

				} else {
					fileConfs[key] = confInfo
				}
			} else {
				fileConfs[key] = confInfo
			}
		}
	}

	// create add/create new config
	fileConfs, events = fSource.addOrCreateConf(fileConfs, configs, events, filePath)

	fSource.Configurations = fileConfs

	return events
}

func (fSource *Source) addOrCreateConf(fileConfs map[string]*ConfigInfo, newconf map[string]interface{},
	events []*event.Event, filePath string) (map[string]*ConfigInfo, []*event.Event) {
	for key, value := range newconf {
		handled := false

		_, ok := fileConfs[key]
		if ok {
			handled = true
		}

		if !handled {
			events = append(events, &event.Event{EventSource: FileConfigSourceConst, Key: key,
				EventType: event.Create, Value: value})
			fileConfs[key] = &ConfigInfo{
				FilePath: filePath,
				Value:    value,
			}
		}
	}

	return fileConfs, events
}

//Cleanup clear all configs
func (fSource *Source) Cleanup() error {
	fSource.filelock.Lock()
	defer fSource.filelock.Unlock()

	if fSource.watchPool != nil && fSource.watchPool.watcher != nil {
		fSource.watchPool.watcher.Close()
	}

	fSource.files = make([]file, 0)
	fSource.Configurations = make(map[string]*ConfigInfo, 0)
	return nil
}

//AddDimensionInfo  is none function
func (fSource *Source) AddDimensionInfo(labels map[string]string) error {
	return nil
}

//Set no use
func (fSource *Source) Set(key string, value interface{}) error {
	return nil
}

//Delete no use
func (fSource *Source) Delete(key string) error {
	return nil
}
