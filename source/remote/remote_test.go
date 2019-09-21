package remote_test

import (
	"github.com/go-chassis/go-archaius/event"
	_ "github.com/go-chassis/go-chassis-config/configcenter"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/stretchr/testify/assert"

	"errors"
	"github.com/go-chassis/go-chassis-config"
	"testing"
	"time"
)

type EventHandler struct {
	EventName  string
	EventKey   string
	EventValue interface{}
}

func (ccenter *EventHandler) OnEvent(event *event.Event) {
	ccenter.EventName = event.EventType
	ccenter.EventKey = event.Key
	ccenter.EventValue = event.Value
}

func TestGetConfigurationsForInvalidCCIP(t *testing.T) {
	opts := config.Options{
		Labels: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		ServerURI:  "http://",
	}
	cc, err := config.NewClient("config_center", opts)
	assert.NoError(t, err)
	ccs := remote.NewConfigCenterSource(cc, 1,
		1)

	_, er := ccs.GetConfigurations()
	if er != nil {
		assert.Error(t, er)
	}

	time.Sleep(2 * time.Second)
	ccs.Cleanup()
}

func TestGetConfigurationsWithCCIP(t *testing.T) {
	opts := config.Options{
		Labels: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		ServerURI:  "http://",
		TenantName: "default",
	}
	cc, err := config.NewClient("config_center", opts)
	assert.NoError(t, err)
	ccs := remote.NewConfigCenterSource(cc, 1, 1)

	t.Log("Accessing concenter source configurations")
	time.Sleep(2 * time.Second)
	_, er := ccs.GetConfigurations()
	if er != nil {
		assert.Error(t, er)
	}
	archaius.Init()
	t.Log("concenter source adding to the archaiuscleanup")
	e := archaius.AddSource(ccs)
	if e != nil {
		assert.Error(t, e)
	}

	t.Log("verifying configcenter configurations by Configs method")
	_, err = ccs.GetConfigurationByKey("refreshInterval")
	if err != nil {
		assert.Error(t, err)
	}

	_, err = ccs.GetConfigurationByKey("test")
	if err != nil {
		assert.Error(t, err)
	}

	t.Log("verifying configcenter name")
	sourceName := ccs.GetSourceName()
	if sourceName != "ConfigCenterSource" {
		t.Error("config-center name is mismatched")
	}

	t.Log("verifying configcenter priority")
	priority := ccs.GetPriority()
	if priority != 0 {
		t.Error("config-center priority is mismatched")
	}

	t.Log("concenter source cleanup")
	ccs.Cleanup()

}

func Test_DynamicConfigHandler(t *testing.T) {
	opts := config.Options{
		Labels: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		ServerURI:  "http://",
	}
	cc, err := config.NewClient("config_center", opts)
	assert.NoError(t, err)
	ccs := remote.NewConfigCenterSource(cc, 1, 1)

	eh := new(EventHandler)

	ccs.Watch(eh)

	//post the new key, or update the already existing key, or delete the existing key to get the events
	time.Sleep(4 * time.Second)

	if eh.EventName == "" {
		err := errors.New("failed to get the event if key is created or updated or deleted")
		assert.Error(t, err)
	}

}

func Test_OnReceive(t *testing.T) {
	opts := config.Options{
		Labels: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		ServerURI:  "http://",
	}
	cc, err := config.NewClient("config_center", opts)
	assert.NoError(t, err)
	ccs := remote.NewConfigCenterSource(cc, 1, 1)

	_, er := ccs.GetConfigurations()
	assert.Error(t, er)

}
