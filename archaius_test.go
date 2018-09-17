package archaius_test

import (
	_ "github.com/go-chassis/go-chassis/initiator"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/core"
	"github.com/go-chassis/go-chassis/core/lager"
	"github.com/stretchr/testify/assert"
)

type EListener struct{}

func (e EListener) Event(event *core.Event) {
	lager.Logger.Infof("config value after change ", event.Key, " | ", event.Value)
}

var filename2 string
var filename3 string

func TestNew(t *testing.T) {
	f1Bytes := []byte(`
age: 14
name: peter
`)
	f2Bytes := []byte(`
addr: 14
number: 1
`)
	f3Bytes := []byte(`
addr: 15
number: 2
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f1.yaml")
	filename2 = filepath.Join(d, "f2.yaml")
	filename3 = filepath.Join(d, "f3.yaml")
	os.Remove(filename1)
	os.Remove(filename2)
	os.Remove(filename3)
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	defer f1.Close()
	f2, err := os.Create(filename2)
	assert.NoError(t, err)
	defer f2.Close()
	f3, err := os.Create(filename3)
	assert.NoError(t, err)
	defer f3.Close()
	_, err = io.WriteString(f1, string(f1Bytes))
	t.Log(string(f1Bytes))
	assert.NoError(t, err)
	_, err = io.WriteString(f2, string(f2Bytes))
	assert.NoError(t, err)
	_, err = io.WriteString(f3, string(f3Bytes))
	assert.NoError(t, err)

	err = archaius.Init(
		archaius.WithRequiredFiles([]string{filename1}),
		archaius.WithOptionalFiles([]string{filename2}),
	)
	assert.NoError(t, err)

}
func TestAddFile(t *testing.T) {
	s := archaius.Get("number")
	s2 := archaius.Get("age")
	assert.Equal(t, 14, s2)
	assert.Equal(t, 1, s)

}
func TestConfig_Get(t *testing.T) {
	s := archaius.Get("age")
	n := archaius.Get("name")
	assert.Equal(t, 14, s)
	assert.Equal(t, "peter", n)
}
func TestConfig_GetInt(t *testing.T) {
	s := archaius.Get("age")
	assert.Equal(t, 14, s)
}
func TestConfig_RegisterListener(t *testing.T) {
	eventHandler := EListener{}
	err := archaius.RegisterListener(eventHandler, "a*")
	assert.NoError(t, err)
	defer archaius.UnRegisterListener(eventHandler, "a*")

}

func TestAddFile_WithFunc(t *testing.T) {
	err := archaius.AddFile(filename3, archaius.WithFileHandler)
	assert.Nil(t, err)
	value := archaius.Get(filename3)
	assert.NotNil(t, value)
	v, err := archaius.WithFileHandler(filename3)
	assert.Nil(t, err)
	for _, j := range v {
		assert.Equal(t, "\naddr: 15\nnumber: 2\n", j)
	}
}
