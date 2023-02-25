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

/*
* Created by on 2017/6/22.
 */
package filesource_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arielsrv/go-archaius/event"
	filesource "github.com/arielsrv/go-archaius/source/file"
	"github.com/stretchr/testify/assert"
)

type TestDynamicConfigHandler struct {
	EventName  string
	EventKey   string
	EventValue interface{}
}

func (t *TestDynamicConfigHandler) OnModuleEvent(events []*event.Event) {
	fmt.Println("implement me")
}

func (t *TestDynamicConfigHandler) OnEvent(e *event.Event) {
	t.EventKey = e.Key
	t.EventName = e.EventType
	t.EventValue = e.Value
}

// GetWorkDir is a function used to get the working directory
func GetWorkDir() (string, error) {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return wd, nil
}
func TestNewFileSource(t *testing.T) {
	root, _ := GetWorkDir()

	file1 := filepath.Join(root, "test1.yaml")
	file2 := filepath.Join(root, "test2.yaml")

	f1content := []byte(`
name: peter
age: 12
`)
	f1content2 := []byte(`
name: peter
age: 13
`)
	f2content := []byte(`
others:
  addr: beijing
  phone: 123
`)
	f1, err := os.Create(file1)
	assert.NoError(t, err)
	f2, err := os.Create(file2)
	assert.NoError(t, err)
	defer f1.Close()
	defer f2.Close()
	defer os.Remove(file1)
	defer os.Remove(file2)

	_, err = io.WriteString(f1, string(f1content))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, string(f2content))
	assert.NoError(t, err)

	fSource := filesource.NewFileSource()
	fSource.Watch(new(TestDynamicConfigHandler))
	//Configuration file1 is adding to the filesource
	err = fSource.AddFile(file1, 0, nil)
	assert.NoError(t, err)
	t.Run("add duplicated file", func(t *testing.T) {
		err = fSource.AddFile(file1, 0, nil)
		assert.NoError(t, err)
	})
	t.Run("add no existing file", func(t *testing.T) {
		err = fSource.AddFile("/notexistingdir/notexisting.yaml", 0, nil)
		assert.Error(t, err)
	})

	//Adding directory to the filesource
	err = fSource.AddFile(filepath.Join(root, "dir"), 0, nil)
	assert.Error(t, err)
	t.Run("check config map", func(t *testing.T) {
		configMap, err := fSource.GetConfigurations()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(configMap))
	})

	t.Run("get name", func(t *testing.T) {
		name, err := fSource.GetConfigurationByKey("name")
		assert.NoError(t, err)
		assert.Equal(t, "peter", name)
	})
	t.Run("get age", func(t *testing.T) {
		age, err := fSource.GetConfigurationByKey("age")
		assert.NoError(t, err)
		assert.Equal(t, 12, age)
	})
	_, err = f1.Write(f1content2)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	t.Run("get age after event", func(t *testing.T) {
		age, err := fSource.GetConfigurationByKey("age")
		assert.NoError(t, err)
		assert.Equal(t, 13, age)
	})

	t.Run("clean up", func(t *testing.T) {
		err := fSource.Cleanup()
		assert.NoError(t, err)
		age, err := fSource.GetConfigurationByKey("age")
		assert.Equal(t, nil, age)
	})
}
