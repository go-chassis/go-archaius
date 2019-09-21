package archaius_test

import (
	"github.com/go-chassis/go-archaius/event"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chassis/go-archaius"
	"github.com/go-mesh/openlogging"
	"github.com/stretchr/testify/assert"
)

type EListener struct{}

func (e EListener) Event(event *event.Event) {
	openlogging.GetLogger().Infof("config value after change ", event.Key, " | ", event.Value)
}

var filename2 string

func TestInit(t *testing.T) {
	f1Bytes := []byte(`
age: 14
name: peter
`)
	f2Bytes := []byte(`
addr: somewhere
number: 1
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f1.yaml")
	filename2 = filepath.Join(d, "f2.yaml")
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(filename1)
	f2, err := os.Create(filename2)
	assert.NoError(t, err)
	defer f2.Close()
	defer os.Remove(filename2)
	_, err = io.WriteString(f1, string(f1Bytes))
	t.Log(string(f1Bytes))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, string(f2Bytes))
	assert.NoError(t, err)
	os.Setenv("age", "15")
	err = archaius.Init(
		archaius.WithRequiredFiles([]string{filename1}),
		archaius.WithOptionalFiles([]string{filename2}),
		archaius.WithENVSource(),
		archaius.WithMemorySource())
	assert.NoError(t, err)
	assert.Equal(t, "15", archaius.Get("age"))
	t.Run("add mem config", func(t *testing.T) {
		archaius.Set("age", "16")
		assert.Equal(t, "16", archaius.Get("age"))
	})
	t.Run("delete mem config", func(t *testing.T) {
		archaius.Delete("age")
		assert.Equal(t, "15", archaius.Get("age"))
	})

}
func TestAddFile(t *testing.T) {
	s := archaius.Get("number")
	assert.Equal(t, 1, s)

}
func TestConfig_Get(t *testing.T) {
	n := archaius.Get("name")
	assert.Equal(t, "peter", n)
}
func TestConfig_GetInt(t *testing.T) {
	s := archaius.Get("age")
	assert.Equal(t, "15", s)
}
func TestConfig_RegisterListener(t *testing.T) {
	eventHandler := EListener{}
	err := archaius.RegisterListener(eventHandler, "a*")
	assert.NoError(t, err)
	defer archaius.UnRegisterListener(eventHandler, "a*")

}
func TestInitConfigCenter(t *testing.T) {
	err := archaius.EnableRemoteSource(&archaius.RemoteInfo{}, nil)
	assert.Error(t, err)
	err = archaius.EnableRemoteSource(&archaius.RemoteInfo{
		ClientType: "fake",
	}, nil)
	assert.Error(t, err)
}
func TestClean(t *testing.T) {
	err := archaius.Clean()
	assert.NoError(t, err)
	s := archaius.Get("age")
	assert.Equal(t, nil, s)

}
