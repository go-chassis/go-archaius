/*
 * Copyright 2019 Huawei Technologies Co., Ltd
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

package configmapource

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arielsrv/go-archaius/event"
	"github.com/stretchr/testify/assert"
)

type EListener struct {
	Name      string
	EventName string
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

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

func TestDynamicConfigurations(t *testing.T) {

	root, _ := GetWorkDir()
	os.Setenv("CHASSIS_HOME", root)

	tmpdir := filepath.Join(root, "tmp")
	filename1 := filepath.Join(root, "tmp", "test1.yaml")
	filename2 := filepath.Join(root, "tmp", "test2.yaml")
	filename3 := filepath.Join(root, "tmp", "test3.yaml")
	filename4 := filepath.Join(root, "tmp", "test4.yaml")
	filename5 := filepath.Join(root, "tmp", "test5.yaml")

	yamlContent1 := "yamlkeytest11: test11\n \nyamlkeytest12: test12\n \nyamlkeytest123: test1231"
	yamlContent2 := "yamlkeytest21: test21\n \nyamlkeytest22: test22\n \nyamlkeytest123: test1232"
	yamlContent3 := "yamlkeytest31: test31\n \nyamlkeytest32: test32\n \nyamlkeytest123: test1233"
	yamlContent4 := "yamlkeytest41: test41\n \nyamlkeytest42: test32\n \nyamlkeytest45: test454"
	yamlContent5 := "yamlkeytest51: test51\n \nyamlkeytest52: test52\n \nyamlkeytest123: test1233"

	os.Remove(filename1)
	os.Remove(filename2)
	os.Remove(filename3)
	os.Remove(filename4)
	os.Remove(filename5)
	os.Remove(tmpdir)
	err := os.Mkdir(tmpdir, 0777)
	check(err)
	defer os.Remove(tmpdir)

	f1, err := os.Create(filename1)
	check(err)
	defer f1.Close()
	defer os.Remove(filename1)
	f2, err := os.Create(filename2)
	check(err)
	defer f2.Close()
	defer os.Remove(filename2)
	f3, err := os.Create(filename3)
	check(err)
	defer f3.Close()
	defer os.Remove(filename3)
	f4, err := os.Create(filename4)
	check(err)
	defer f4.Close()
	defer os.Remove(filename4)
	f5, err := os.Create(filename5)
	check(err)
	defer f5.Close()
	defer os.Remove(filename5)

	_, err = io.WriteString(f1, yamlContent1)
	check(err)
	_, err = io.WriteString(f2, yamlContent2)
	check(err)
	_, err = io.WriteString(f3, yamlContent3)
	check(err)
	_, err = io.WriteString(f4, yamlContent4)
	check(err)
	_, err = io.WriteString(f5, yamlContent5)
	check(err)

	cmSource := NewConfigMapSource()
	cmSource.AddFile(filename1, 0, nil)
	cmSource.AddFile(filename2, 1, nil)
	cmSource.AddFile(filename3, 2, nil)

	dynHandler := new(TestDynamicConfigHandler)
	cmSource.Watch(dynHandler)
	time.Sleep(1 * time.Second)

	t.Log("generate event by inserting some value into file")
	yamlContent1 = "\nyamlkeytest13: test13\n"
	_, err = io.WriteString(f1, yamlContent1)
	check(err)
	time.Sleep(10 * time.Millisecond)

	t.Log("Verifying the key of highest priority file(filename1)")
	configkey, err := cmSource.GetConfigurationByKey("yamlkeytest13")
	if configkey != "test13" {
		t.Error("Failed to get the latest event key value pair")
	}

	//Accessing key of file2 priority is 1
	configkey, _ = cmSource.GetConfigurationByKey("yamlkeytest21")
	if configkey != "test21" {
		t.Error("Failed to get the latest event key value pair")
	}

	//verifying the key of highest priority file(filename1)
	configkey, _ = cmSource.GetConfigurationByKey("yamlkeytest123")
	if configkey != "test1231" {
		t.Error("Failed to get the latest event key value pair")
	}

	//generating the key from highest priority file(filename1)
	yamlContent1 = "\nyamlkeytest123: test12311\n"
	_, err = io.WriteString(f1, yamlContent1)
	check(err)
	time.Sleep(10 * time.Millisecond)

	//Verifying the of highest priority file(filename1)
	configkey, err = cmSource.GetConfigurationByKey("yamlkeytest123")
	if configkey != "test12311" {
		t.Error("filesource updating the key from lowest priority file")
	}

	t.Log("generating the key from lowest priority file(filename3)")
	yamlContent3 = "\nyamlkeytest123: test12333\n"
	_, err = io.WriteString(f3, yamlContent3)
	check(err)
	time.Sleep(10 * time.Millisecond)

	t.Log("verifying the key of lowest priority file(filename3)")
	configkey, err = cmSource.GetConfigurationByKey("yamlkeytest123")
	if configkey == "test12333" {
		t.Error("filesource updating the key from lowest priority file")
	}

	t.Log("adding new files after dynhandler is inited")
	cmSource.AddFile(filename4, 3, nil)
	cmSource.AddFile(filename5, 4, nil)
	time.Sleep(10 * time.Millisecond)

	t.Log("verifying the configurations of newely added files")
	configkey, err = cmSource.GetConfigurationByKey("yamlkeytest41")
	assert.Equal(t, "test41", configkey)
	configkey, _ = cmSource.GetConfigurationByKey("yamlkeytest51")
	assert.Equal(t, "test51", configkey)

	t.Log("creating the event from newely added file(filename4)")
	yamlContent4 = "\nyamlkeytest45: test454\n"
	_, err = io.WriteString(f4, yamlContent4)
	check(err)
	time.Sleep(10 * time.Millisecond)
	configkey, _ = cmSource.GetConfigurationByKey("yamlkeytest45")
	assert.Equal(t, "test454", configkey)

	t.Log("update event from lowest priority file(filename5)")
	yamlContent5 = "\nyamlkeytest45: test455\n"
	_, err = io.WriteString(f5, yamlContent5)
	check(err)
	time.Sleep(10 * time.Millisecond)
	configkey, _ = cmSource.GetConfigurationByKey("yamlkeytest45")
	t.Log("verifying the event from lowest priority file(filename5)")
	assert.NotEqual(t, "test455", configkey)
	assert.Equal(t, "test454", configkey)

	t.Log("filesource cleanup")
	configsourcecleanup := cmSource.Cleanup()
	if configsourcecleanup != nil {
		t.Error("filesource cleanup is Failed")
	}
}

// delete old directory and create new directory
func TestConfigMapSource2(t *testing.T) {

	root, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Error("get file error")
	}

	t.Log("Test configmapsource.go")

	confmapDir := filepath.Join(root, "configmap")
	dir1 := filepath.Join(root, "configmap", "dir1")
	file1 := filepath.Join(root, "configmap/dir1", "test1.yaml")
	dir2 := filepath.Join(root, "configmap", "dir2")
	file2 := filepath.Join(root, "configmap/dir2", "test2.yaml")
	dir3 := filepath.Join(root, "configmap", "dir3")
	file3 := filepath.Join(root, "configmap/dir3", "test3.yaml")

	f1content := "NAME11: test11\n \nNAME12: test12"
	f2content := "NAME21: test21\n \nNAME22: test22"
	f3content := "NAME21: test31\n \nNAME22: test32"

	os.Remove(file1)
	os.Remove(dir1)

	os.Remove(file2)
	os.Remove(dir2)

	os.Remove(file3)
	os.Remove(dir3)

	os.Remove(confmapDir)

	err = os.Mkdir(confmapDir, 0777)
	check(err)
	defer os.Remove(confmapDir)

	err = os.Mkdir(dir1, 0777)
	check(err)
	defer os.Remove(dir1)
	f1, err := os.Create(file1)
	check(err)
	defer f1.Close()
	defer os.Remove(file1)

	_, err = io.WriteString(f1, string(f1content))
	check(err)

	cmSource := NewConfigMapSource()

	dynHandler := new(TestDynamicConfigHandler)
	cmSource.Watch(dynHandler)
	time.Sleep(1 * time.Second)

	err = cmSource.AddFile(confmapDir, 0, nil)
	if err != nil {
		t.Error(err)
	}

	t.Log("verifying configmapsource configurations by Configs method")
	_, err = cmSource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get the configurations from configmapsource")
	}

	configkey, _ := cmSource.GetConfigurationByKey("NAME11")
	if configkey != "test11" {
		t.Error("Failed to the configmapsource keyvalue pair")
	}

	//create directory and files
	t.Log("create new directory")
	err = os.Mkdir(dir2, 0777)
	check(err)
	f2, err := os.Create(file2)
	check(err)
	_, err = io.WriteString(f2, f2content)
	check(err)

	time.Sleep(1 * time.Second)

	_, err = cmSource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get the configurations from configmapsource")
	}

	configkey, _ = cmSource.GetConfigurationByKey("NAME21")
	if configkey != "test21" {
		t.Error("Failed to the configmapsource keyvalue pair")
	}

	//delete directory2 and files, create directory3 and files
	t.Log("delete old directory and create new directory")
	os.Remove(file2)
	f2.Close()
	os.Remove(dir2)
	err = os.Mkdir(dir3, 0777)
	check(err)
	defer os.Remove(dir3)
	f3, err := os.Create(file3)
	check(err)
	defer f3.Close()
	defer os.Remove(file3)
	_, err = io.WriteString(f3, f3content)
	check(err)

	time.Sleep(1 * time.Second)

	_, err = cmSource.GetConfigurations()
	if err != nil {
		t.Error("Failed to get the configurations from configmapsource")
	}

	configkey, _ = cmSource.GetConfigurationByKey("NAME21")
	if configkey != "test31" {
		t.Error("Failed to the configmapsource keyvalue pair")
	}

	configsourcecleanup := cmSource.Cleanup()
	if configsourcecleanup != nil {
		t.Error("configmapsource cleanup is Failed")
	}

}
