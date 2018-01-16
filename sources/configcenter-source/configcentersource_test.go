package configcentersource

import (
	"github.com/ServiceComb/go-cc-client/member-discovery"

	"github.com/ServiceComb/go-archaius/core"
	"github.com/ServiceComb/go-chassis/core/archaius"
	"github.com/ServiceComb/go-chassis/core/config"
	"github.com/ServiceComb/go-chassis/core/config/model"
	"github.com/stretchr/testify/assert"

	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"testing"
	"time"
)

type Testingsource struct {
	configuration  map[string]interface{}
	changeCallback core.DynamicConfigCallback
}

type TestDynamicConfigHandler struct {
	EventName  string
	EventKey   string
	EventValue interface{}
}

func (ccenter *TestDynamicConfigHandler) OnEvent(event *core.Event) {

	ccenter.EventName = event.EventType
	ccenter.EventKey = event.Key
	ccenter.EventValue = event.Value
}

func (*Testingsource) GetDimensionInfo() string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 5)

	for i := 0; i < 5; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}

	dimensioninfo := string(result)
	return dimensioninfo
}

func (*Testingsource) GetConfigServer() []string {
	configserver := []string{`http://10.18.206.218:30103`}

	return configserver
}

func (*Testingsource) GetInvalidConfigServer() []string {
	return nil
}

func TestGetConfigurationsForInvalidCCIP(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	os.Setenv("CHASSIS_HOME", gopath+"/src/code.huawei.com/cse/go-chassis-examples/discovery/server/")
	config.Init()
	testSource := &Testingsource{}

	t.Log("Test configcentersource.go")

	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.Config.Client.APIVersion.Version = "v2"
	memDiscovery := memberdiscovery.NewConfiCenterInit(nil, "default", false)
	memDiscovery.ConfigurationInit(testSource.GetInvalidConfigServer())
	configcentersource := NewConfigCenterSource(memDiscovery, testSource.GetDimensionInfo(), nil, "default", 1, 1, false)

	_, er := configcentersource.GetConfigurations()
	if er != nil {
		assert.Error(t, er)
	}

	time.Sleep(2 * time.Second)
	configCenterConfig.Cleanup()
}

func TestGetConfigurationsWithCCIP(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	os.Setenv("CHASSIS_HOME", gopath+"/src/code.huawei.com/cse/go-chassis-examples/discovery/server/")
	config.Init()
	testSource := &Testingsource{}

	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.Config.Client.APIVersion.Version = "v2"
	memDiscovery := memberdiscovery.NewConfiCenterInit(nil, "default", false)
	memDiscovery.ConfigurationInit(testSource.GetConfigServer())
	configcentersource := NewConfigCenterSource(memDiscovery, testSource.GetDimensionInfo(), nil, "default", 1, 1, false)
	config.GlobalDefinition = &model.GlobalCfg{}

	t.Log("Accessing concenter source configurations")
	time.Sleep(2 * time.Second)
	_, er := configcentersource.GetConfigurations()
	if er != nil {
		assert.Error(t, er)
	}
	archaius.Init()
	t.Log("concenter source adding to the archaiuscleanup")
	e := archaius.DefaultConf.ConfigFactory.AddSource(configcentersource)
	if e != nil {
		assert.Error(t, e)
	}

	t.Log("verifying configcentersource configurations by GetConfigurations method")
	_, err := configcentersource.GetConfigurationByKey("refreshInterval")
	if err != nil {
		assert.Error(t, err)
	}

	_, err = configcentersource.GetConfigurationByKey("test")
	if err != nil {
		assert.Error(t, err)
	}

	_, err = configcentersource.GetConfigurationByKeyAndDimensionInfo("data@default#0.1", "test")
	if err != nil {
		assert.Error(t, err)
	}

	t.Log("verifying configcentersource name")
	sourceName := configCenterConfig.GetSourceName()
	if sourceName != "ConfigCenterSource" {
		t.Error("config-center name is mismatched")
	}

	t.Log("verifying configcentersource priority")
	priority := configCenterConfig.GetPriority()
	if priority != 0 {
		t.Error("config-center priority is mismatched")
	}

	t.Log("concenter source cleanup")
	configCenterConfig.Cleanup()

}

func Test_DynamicConfigHandler(t *testing.T) {
	testsource := &Testingsource{}

	config.Init()
	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.Config.Client.APIVersion.Version = "v2"
	memDiscovery := memberdiscovery.NewConfiCenterInit(nil, "default", false)
	memDiscovery.ConfigurationInit(testsource.GetConfigServer())
	configcentersource := NewConfigCenterSource(memDiscovery, testsource.GetDimensionInfo(), nil, "default", 1, 1, false)

	dynamicconfig := new(TestDynamicConfigHandler)

	configcentersource.DynamicConfigHandler(dynamicconfig)

	//post the new key, or update the already existing key, or delete the existing key to get the events
	time.Sleep(4 * time.Second)

	if dynamicconfig.EventName == "" {
		err := errors.New("Failed to get the event if key is created or updated or deleted")
		assert.Error(t, err)
	}

}

func Test_OnReceive(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	os.Setenv("CHASSIS_HOME", gopath+"/src/code.huawei.com/cse/go-chassis-examples/discovery/server/")
	config.Init()
	testSource := &Testingsource{}

	config.GlobalDefinition = &model.GlobalCfg{}
	config.GlobalDefinition.Cse.Config.Client.APIVersion.Version = "v2"
	memDiscovery := memberdiscovery.NewConfiCenterInit(nil, "default", false)
	memDiscovery.ConfigurationInit(testSource.GetInvalidConfigServer())
	configcentersource := NewConfigCenterSource(memDiscovery, testSource.GetDimensionInfo(), nil, "default", 0, 1, false)

	_, er := configcentersource.GetConfigurations()
	if er != nil {
		assert.Error(t, er)
	}

	dynamicconfig := new(TestDynamicConfigHandler)

	configCenterEvent := new(ConfigCenterEvent)
	configCenterEvent.Action = "test"
	check := make(map[string]interface{})
	check["refreshMode"] = 7
	data, _ := json.Marshal(&check)
	configCenterEvent.Value = string(data)

	data1, _ := json.Marshal(&configCenterEvent)
	configCenterEventHandler := new(ConfigCenterEventHandler)
	configCenterEventHandler.configSource = configCenterConfig
	configCenterEventHandler.callback = dynamicconfig

	configCenterEventHandler.OnReceive(data1)
}
