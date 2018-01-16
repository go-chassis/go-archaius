package configmanager

import (
	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-archaius/core/event-system"
	"github.com/ServiceComb/go-archaius/sources/commandline-source"
	"github.com/ServiceComb/go-archaius/sources/external-source"
	"github.com/ServiceComb/go-archaius/sources/file-source"
	"github.com/ServiceComb/go-archaius/sources/test-source"
	"github.com/ServiceComb/go-chassis/core/config/model"
	"github.com/ServiceComb/go-chassis/core/lager"
	"github.com/ServiceComb/go-chassis/util/fileutil"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func populateCmdConfig() {
	os.Args = append(os.Args, "--testcmdkey1=cmdkey1")
	os.Args = append(os.Args, "--testcmdkey2=cmdkey2")
	os.Args = append(os.Args, "--aaa=cmdkey3")
}

func TestConfigurationManager(t *testing.T) {

	testConfig := map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := testsource.NewTestSource(testConfig)

	dispatcher := eventsystem.NewDispatcher()
	confmanager := NewConfigurationManager(dispatcher)
	t.Log("Test configurationmanager.go")

	//supplying nil source
	var ab core.ConfigSource
	lager.Initialize("", "INFO", "", "size", true, 1, 10, 7)
	err := confmanager.AddSource(ab, DefaultPriority)
	if err == nil {
		t.Error("Failed to identify invalid or nil source")
	}

	//note: lowest value has highest priority
	//testSource priority 	=	0
	//commandlinePriority 	= 	1
	//envSourcePriority 	= 	2
	//fileSourcePriority    = 	3
	//extSourcePriority 	= 	4

	t.Log("Adding testSource to the configuration manager")
	err = confmanager.AddSource(testSource, testSource.GetPriority())
	if err != nil {
		t.Error("Error in adding testSource to the  configuration manager", err)
	}

	//supplying duplicate source
	err = confmanager.AddSource(testSource, testSource.GetPriority())
	if err == nil {
		t.Error("Failed to identify the duplicate config source")
	}

	t.Log("verifying the key testsource key existence from configmanager")
	configvalue := confmanager.GetConfigurationsByKey("aaa")
	if configvalue != "111" {
		t.Error("Failed to get the existing keyvalue pair from configmanager")
	}

	//getting command line configurations
	populateCmdConfig()
	cmdlinesource := commandlinesource.NewCommandlineConfigSource()

	t.Log("Adding cmdlinesource to the configuration manager")
	err = confmanager.AddSource(cmdlinesource, cmdlinesource.GetPriority())
	if err != nil {
		t.Error("Error in adding cmdlinesource to the  configuration manager", err)
	}

	t.Log("Verifying the key of lowest priority(cmdline) source")
	configvalue = confmanager.GetConfigurationsByKey("aaa")
	if configvalue == "cmdkey3" {
		t.Error("Failed to get the keyvalue pair of the highest priority source from configmanager")
	}

	//Accessing not existing key in configmanager
	configvalue = confmanager.GetConfigurationsByKey("notExistingKey")
	if configvalue != nil {
		t.Error("configmanager having invalidkeys")
	}

	t.Log("accessing all configurations")
	configurations := confmanager.GetConfigurations()
	if configurations["testcmdkey1"] != "cmdkey1" && configurations["bbb"] != "222" {
		t.Error("Failed to get configurations")
	}

	confByDI, _ := confmanager.GetConfigurationsByDimensionInfo("darklaunch@default#0.0.1")
	assert.NotEqual(t, confByDI, nil)

	addDI, _ := confmanager.AddDimensionInfo("testdi@default")
	assert.NotEqual(t, addDI, nil)

	t.Log("create event through testsource")
	time.Sleep(10 * time.Millisecond)
	testsource.AddConfig("zzz", "333")
	time.Sleep(10 * time.Millisecond)
	t.Log("Accessing keyvalue pair of created event")
	configvalue = confmanager.GetConfigurationsByKey("zzz")
	if configvalue != "333" {
		t.Error("Failed to get the keyvalue pair for created event")
	}

	t.Log("Refresh the testSource configurations")
	testsource.AddConfig("ccc", "444")
	err = confmanager.Refresh(testSource.GetSourceName())
	if err != nil {
		t.Error(err)
	}

	t.Log("verifying the configurations after updation")
	configurations = confmanager.GetConfigurations()
	if configurations["ccc"] != "444" {
		t.Error("Failed to refresh the configurations")
	}

	//Verifying with the invalidsource refreshing
	if err = confmanager.Refresh("InvalidSource"); err == nil {
		t.Error(err)
	}

	//Supplying nil event
	ConfManager2 := &ConfigurationManager{}
	var event *core.Event = nil
	ConfManager2.OnEvent(event)

	t.Log("Adding external source to generate the events based on priority of the key")
	extsource := externalconfigsource.NewExternalConfigurationSource()
	confmanager.AddSource(extsource, extsource.GetPriority())

	t.Log("Create event through extsource")
	extsource.AddKeyValue("Commonkey", "extsource")
	time.Sleep(10 * time.Millisecond)
	if configvalue = confmanager.GetConfigurationsByKey("Commonkey"); configvalue != "extsource" {
		t.Error("Failed to get the create event of extsource from configmanager")
	}

	t.Log("update event through testsource(highest priority)")
	testsource.AddConfig("Commonkey", "testsource")
	time.Sleep(10 * time.Millisecond)
	configvalue = confmanager.GetConfigurationsByKey("Commonkey")
	if configvalue != "testsource" {
		t.Error("Failed to get the update event of highest priority source")
	}

	t.Log("update event through extsource(lowest priority)")
	extsource.AddKeyValue("Commonkey", "extsource2")
	time.Sleep(10 * time.Millisecond)
	configvalue = confmanager.GetConfigurationsByKey("Commonkey")
	if configvalue == "extsource2" {
		t.Error("key is updaing from lowest priority source")
	}

	t.Log("update event through testsource(highest priority)")
	testsource.AddConfig("Commonkey", "testsource2")
	time.Sleep(10 * time.Millisecond)
	configvalue = confmanager.GetConfigurationsByKey("Commonkey")
	if configvalue != "testsource2" {
		t.Error("Failed to get the update event of highest priority source")
	}

	t.Log("checking the functionality of IsKeyExist")
	configkey3 := confmanager.IsKeyExist("Commonkey")
	configkey4 := confmanager.IsKeyExist("notexistingkey")
	if configkey3 != true && configkey4 != false {
		t.Error("Failed to identify the status of the keys")
	}

	configmap := make(map[string]interface{}, 0)
	err = confmanager.Unmarshal(&configmap)
	if err != nil {
		t.Error("Failed to unmarshal map: ", err)
	}

	if configmap["Commonkey"] != "testsource2" || configmap["testcmdkey2"] != "cmdkey2" {
		t.Error("Failed to get the keyvalue pairs through unmarshall")
	}

	var testobj string
	err = confmanager.Unmarshal(testobj)
	if err == nil {
		t.Error("Failed to detect invalid object while unmarshalling")
	}

	testsource.CleanupTestSource()
	confmanager.Cleanup()

}

func TestConfigurationManager_AddSource(t *testing.T) {

	file := []byte(`
region:
  name: us-east
  availableZone: us-east-1
APPLICATION_ID: CSE
register_type: servicecenter
cse:
  loadbalance:
    strategyName: RoundRobin
    retryEnabled: false
    retryOnNext: 2
    retryOnSame: 3
    backoff:
      kind: constant
      minMs: 200
      maxMs: 400
  service:
    registry:
      type: servicecenter
      scope: full
      autodiscovery: false
      address: 10.19.169.119:30100
      #register: manual
      refeshInterval : 30s
      watch: true
  protocols:
    highway:
      listenAddress: 127.0.0.1:8082
      advertiseAddress: 127.0.0.1:8082
      transport: tcp
    rest:
      listenAddress: 127.0.0.1:8083
      advertiseAddress: 127.0.0.1:8083
      transport: tcp
  handler:
    chain:
      provider:
        default: bizkeeper-provider
ssl:
  cipherPlugin: default
  verifyPeer: false
  cipherSuits: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
  protocol: TLSv1.2
  caFile:
  certFile:
  keyFile:
  certPwdFile:

commonkey3 : filesource
`)

	root, _ := fileutil.GetWorkDir()
	os.Setenv("CHASSIS_HOME", root)
	t.Log(os.Getenv("CHASSIS_HOME"))

	tmpdir := filepath.Join(root, "tmp")
	file1 := filepath.Join(root, "tmp", "chassis.yaml")

	os.Remove(file1)
	os.Remove(tmpdir)
	err := os.Mkdir(tmpdir, 0777)
	check(err)
	defer os.Remove(tmpdir)

	f1, err := os.Create(file1)
	check(err)
	defer f1.Close()
	defer os.Remove(file1)
	_, err = io.WriteString(f1, string(file))

	dispatcher := eventsystem.NewDispatcher()
	confmanager := NewConfigurationManager(dispatcher)

	fsource := filesource.NewYamlConfigurationSource()
	fsource.AddFileSource(file1, 0)

	confmanager.AddSource(fsource, fsource.GetPriority())
	time.Sleep(2 * time.Second)

	t.Log("verifying Unmarshalling")
	globalDef := model.GlobalCfg{}
	err = confmanager.Unmarshal(&globalDef)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "CSE", globalDef.AppID)
	assert.Equal(t, 2, globalDef.Cse.Loadbalance.RetryOnNext)
	assert.Equal(t, "default", globalDef.Ssl["cipherPlugin"])
	assert.Equal(t, "us-east", globalDef.DataCenter.Name)

	err = confmanager.Unmarshal("invalidobject")
	if err == nil {
		t.Error("Failed tp detect the invalid object while unmarshalling")
	}

	Namestring := "String"
	err = confmanager.Unmarshal(&Namestring)
	if err != nil {
		t.Error("Unmarshalling is fail on string object")
	}

	t.Log("verifying the commonkey across the sources ")
	assert.Equal(t, "filesource", confmanager.GetConfigurationsByKey("commonkey3"))

	extsource := externalconfigsource.NewExternalConfigurationSource()
	confmanager.AddSource(extsource, extsource.GetPriority())

	//update the event through extsource
	extsource.AddKeyValue("commonkey3", "extsource")
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, "filesource", confmanager.GetConfigurationsByKey("commonkey3"))
	assert.NotEqual(t, "extsource", confmanager.GetConfigurationsByKey("commonkey3"))

	assert.NotEqual(t, "extsource", confmanager.GetConfigurationsByKeyAndDimensionInfo("data@default#0.1", "commonkey3"))

	// deleting the common key in filesource
	_, err = exec.Command("sed", "-i", "/commonkey3/d", file1).Output()
	assert.Equal(t, nil, err)
	time.Sleep(10 * time.Millisecond)
	assert.NotEqual(t, "filesource", confmanager.GetConfigurationsByKey("commonkey3"))
	assert.Equal(t, "extsource", confmanager.GetConfigurationsByKey("commonkey3"))

	//update the event through extsource
	extsource.AddKeyValue("commonkey3", "extsource2")
	time.Sleep(10 * time.Millisecond)
	assert.NotEqual(t, "filesource", confmanager.GetConfigurationsByKey("commonkey3"))
	assert.Equal(t, "extsource2", confmanager.GetConfigurationsByKey("commonkey3"))

	testConfig := map[string]interface{}{"aaa": "111", "bbb": "222"}
	testSource := testsource.NewTestSource(testConfig)
	err = confmanager.AddSource(testSource, testSource.GetPriority())
	assert.Equal(t, nil, err)
	time.Sleep(10 * time.Millisecond)

	//updating the common key from high priority source(testsource)
	testsource.AddConfig("commonkey3", "testsource")
	time.Sleep(10 * time.Millisecond)
	assert.NotEqual(t, "filesource", confmanager.GetConfigurationsByKey("commonkey3"))
	assert.NotEqual(t, "extsource2", confmanager.GetConfigurationsByKey("commonkey3"))
	assert.Equal(t, "testsource", confmanager.GetConfigurationsByKey("commonkey3"))

	confmanager.Cleanup()

}
