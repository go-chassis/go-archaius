package remote_test

import (
	"github.com/go-chassis/go-archaius/event"
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/stretchr/testify/assert"
	"time"

	"errors"
	"github.com/go-chassis/go-chassis-config"
	"testing"
)

type mockClient struct {
	opts        config.Options
	configsInfo map[string]interface{}
}

// NewClient init the necessary objects needed for seamless communication to Kie Server
func NewClient(options config.Options) (config.Client, error) {
	kieClient := &mockClient{
		opts: options,
		configsInfo: map[string]interface{}{
			"some.enable": true,
		},
	}
	return kieClient, nil
}
func init() {
	config.InstallConfigClientPlugin("mock-client", NewClient)
}

// PullConfigs is used for pull config from servicecomb-kie
func (c *mockClient) PullConfigs(labels ...map[string]string) (map[string]interface{}, error) {

	return c.configsInfo, nil
}

// PullConfig get config by key and labels.
func (c *mockClient) PullConfig(key, contentType string, labels map[string]string) (interface{}, error) {
	return nil, errors.New("can not find value")
}

//PushConfigs put config in kie by key and labels.
func (c *mockClient) PushConfigs(data map[string]interface{}, labels map[string]string) (map[string]interface{}, error) {
	c.configsInfo = data
	return nil, nil
}

//DeleteConfigsByKeys use keyId for delete
func (c *mockClient) DeleteConfigsByKeys(keys []string, labels map[string]string) (map[string]interface{}, error) {
	delete(c.configsInfo, keys[0])
	return nil, nil
}

//Watch not implemented because kie not support.
func (c *mockClient) Watch(f func(map[string]interface{}), errHandler func(err error), labels map[string]string) error {
	// TODO watch change events
	for {
		time.Sleep(1 * time.Second)
		f(c.configsInfo)
	}

	return errors.New("not implemented")
}

//Options.
func (c *mockClient) Options() config.Options {
	return c.opts
}

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

func TestNewConfigCenterSource(t *testing.T) {
	opts := config.Options{
		Labels: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		ServerURI:  "http://",
	}
	cc, err := config.NewClient("mock-client", opts)
	assert.NoError(t, err)
	ccs := remote.NewConfigCenterSource(cc, 1,
		1)

	configs, err := ccs.GetConfigurations()
	assert.NoError(t, err)
	assert.Equal(t, true, configs["some.enable"])

	eh := new(EventHandler)

	_ = ccs.Watch(eh)
	_, _ = cc.PushConfigs(map[string]interface{}{
		"some.enable": true,
		"some":        "new",
	}, nil)

	time.Sleep(2 * time.Second)
	configs, err = ccs.GetConfigurations()
	assert.NoError(t, err)
	assert.Equal(t, "new", configs["some"])
}
