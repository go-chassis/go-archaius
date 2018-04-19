package goarchaius

import (
	"testing"

	"github.com/ServiceComb/go-archaius/core"
	archlager "github.com/ServiceComb/go-archaius/lager"
	"github.com/ServiceComb/go-archaius/sources/file-source"
	"github.com/ServiceComb/go-archaius/sources/memory-source"
	"github.com/ServiceComb/go-archaius/sources/test-source"
	"github.com/ServiceComb/go-chassis/util/fileutil"
	"github.com/ServiceComb/paas-lager/third_party/forked/cloudfoundry/lager"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"time"
)

type ConfigStruct struct {
	Yamltest1 int `yaml:"yamltest1"`
}

type EventListener struct{}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func populateCmdConfig() {
	os.Args = append(os.Args, "--testcmdkey1=cmdkey1")
	os.Args = append(os.Args, "--testcmdkey2=cmdkey2")
	os.Args = append(os.Args, "--commonkey=cmdsource1")
}

func TestConfigFactory(t *testing.T) {

	var lagger lager.Logger
	root, _ := fileutil.GetWorkDir()
	os.Setenv("CHASSIS_HOME", root)
	t.Log(os.Getenv("CHASSIS_HOME"))
	t.Log("Test configurationfactory.go")
	f1content := "APPLICATION_ID: CSE\n  \ncse:\n  service:\n    registry:\n      type: servicecenter\n  protocols:\n       highway:\n         listenAddress: 127.0.0.1:8080\n  \nssl:\n  test.consumer.certFile: test.cer\n  test.consumer.keyFile: test.key\n"

	archlager.InitLager(nil)
	confdir := filepath.Join(root, "conf")
	filename1 := filepath.Join(root, "conf", "chassis.yaml")

	os.Remove(filename1)
	os.Remove(confdir)
	err := os.Mkdir(confdir, 0777)
	check(err)
	f1, err1 := os.Create(filename1)
	check(err1)
	defer os.Remove(confdir)
	defer f1.Close()
	defer os.Remove(filename1)

	_, err1 = io.WriteString(f1, f1content)
	populateCmdConfig()

	_, err = NewConfigFactory(lagger)
	assert.Equal(t, nil, err)

	factory, err := NewConfigFactory(lagger)
	assert.Equal(t, nil, err)

	t.Log("verifying methods before config factory initialization")
	assert.Equal(t, nil, factory.GetValue("testkey"))
	assert.Equal(t, nil, factory.AddSource(nil))
	assert.Equal(t, map[string]interface{}(map[string]interface{}(nil)), factory.GetConfigurations())
	assert.Equal(t, false, factory.IsKeyExist("testkey"))
	assert.Equal(t, nil, factory.Unmarshal("testkey"))
	assert.Equal(t, nil, factory.GetConfigurationByKey("testkey"))
	assert.Equal(t, nil, factory.AddSource(nil))
	assert.Equal(t, nil, factory.GetConfigurationByKeyAndDimensionInfo("data@default#0.1", "hello"))

	factory.DeInit()
	factory.Init()
	defer factory.DeInit()

	//note: lowest value has highest priority
	//testSource priority 	=	0
	//memSourcePriority 	= 	1
	//commandlinePriority 	= 	2
	//envSourcePriority 	= 	3
	//fileSourcePriority    = 	4

	time.Sleep(10 * time.Millisecond)
	eventHandler := EventListener{}
	t.Log("Register Listener")
	err = factory.RegisterListener(eventHandler, "a*")
	if err != nil {
		t.Error(err)
	}
	defer factory.UnRegisterListener(eventHandler, "a*")
	defer t.Log("UnRegister Listener")

	t.Log("verifying existing configuration keyvalue pair")
	configvalue := factory.GetConfigurationByKey("commonkey")
	if configvalue != "cmdsource1" {
		t.Error("Failed to get the existing keyvalue pair")
	}

	t.Log("Adding filesource to the configfactroy")
	fsource := filesource.NewYamlConfigurationSource()
	fsource.AddFileSource(filename1, 0)
	factory.AddSource(fsource)

	t.Log("Generating event through testsource(priority 1)")
	memsource := memoryconfigsource.NewMemoryConfigurationSource()
	memsource.AddKeyValue("commonkey", "memsource1")

	t.Log("verifying the key of lower priority source")
	time.Sleep(10 * time.Millisecond)
	configvalue = factory.GetConfigurationByKey("commonkey")
	if configvalue != "memsource1" {
		t.Error("Failed to get the existing keyvalue pair")
	}

	t.Log("Adding testsource to the configfactory")
	testConfig := map[string]interface{}{"aaa": "111", "bbb": "222", "commonkey": "testsource1"}
	testSource := testsource.NewTestSource(testConfig)
	factory.AddSource(testSource)
	defer testsource.CleanupTestSource()

	t.Log("verifying common configuration keyvalue pair ")
	time.Sleep(10 * time.Millisecond)
	configvalue = factory.GetConfigurationByKey("commonkey")
	if configvalue != "testsource1" {
		t.Error("Failed to get the key highest priority sorce")
	}

	t.Log("verifying filesource configurations ")
	configurations := factory.GetConfigurations()
	if configurations["testcmdkey2"] != "cmdkey2" || configurations["APPLICATION_ID"] != "CSE" {
		t.Error("Failed to get the configurations")
	}

	confByDI := factory.GetConfigurationsByDimensionInfo("darklaunch@default#0.0.1")
	assert.NotEqual(t, confByDI, nil)

	addDI, _ := factory.AddByDimensionInfo("darklaunch@default#0.0.1")
	assert.NotEqual(t, addDI, nil)

	if factory.IsKeyExist("commonkey") != true || factory.IsKeyExist("notexistingkey") != false {
		t.Error("Failed to get the exist status of the keys")
	}

	t.Log("verifying memsource configurations and accessing in different data type formats")
	memsource.AddKeyValue("stringkey", "true")
	time.Sleep(10 * time.Millisecond)
	configvalue2, err := factory.GetValue("stringkey").ToBool()
	if err != nil || configvalue2 != true {
		t.Error("failed to get the value in bool")
	}

	memsource.AddKeyValue("boolkey", "hello")
	time.Sleep(10 * time.Millisecond)
	configvalue3, err := factory.GetValue("boolkey").ToBool()
	if err != nil || configvalue3 != false {
		t.Error("Failed to get the value for string in convertion to bool")
	}

	configvalue4, err := factory.GetValue("nokey").ToBool()
	if err == nil || configvalue4 != false {
		t.Error("Error for nil key and value")
	}

	data, err := factory.GetValueByDI("darklaunch@default#0.0.1", "hi").ToString()
	assert.Equal(t, data, "")
	assert.Error(t, err)

	configmap := make(map[string]interface{}, 0)
	err = factory.Unmarshal(&configmap)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)
	if configmap["testcmdkey1"] != "cmdkey1" || configmap["aaa"] != "111" {
		t.Error("Failed to get the keyvalue pairs through unmarshall")
	}

	//supplying nil listener.
	var listener core.EventListener
	err = factory.RegisterListener(listener, "key")
	if err == nil {
		t.Error("Failed to detect the nil listener while registering")
	}

	//supplying nil listener
	err = factory.UnRegisterListener(listener, "key")
	if err == nil {
		t.Error("Failed to detect the nil listener while unregistering")
	}
}

func (e EventListener) Event(event *core.Event) {
	archlager.Logger.Infof("config value after change ", event.Key, " | ", event.Value)
}
