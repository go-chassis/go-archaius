package archaius_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/arielsrv/go-archaius"
	"github.com/arielsrv/go-archaius/event"
	"github.com/go-chassis/openlog"
	"github.com/stretchr/testify/assert"
)

type EListener struct{}

func (e EListener) Event(event *event.Event) {
	openlog.Info(fmt.Sprintf("config value after change %s |%s", event.Key, event.Value))
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
exist: true
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

}
func TestConfig_Get(t *testing.T) {
	s := archaius.Get("number")
	assert.Equal(t, 1, s)

	e := archaius.GetBool("exist", false)
	assert.Equal(t, true, e)

	n := archaius.Get("name")
	assert.Equal(t, "peter", n)

	n3 := archaius.GetString("name", "")
	assert.Equal(t, "peter", n3)

	b := archaius.Exist("name")
	assert.True(t, b)

	b = archaius.Exist("none")
	assert.False(t, b)

	m := archaius.GetConfigs()
	t.Log(m)
	assert.Equal(t, 1, m["number"])

	// case: view config keys and its original source
	// context: archaius is initialized and has config key: "number", value: 1 from source: FileSource
	// expected result: able to get config key and its original source
	t.Run("get all configs along with source", func(t *testing.T) {
		c1 := archaius.GetConfigsWithSourceNames()
		assert.Equal(t, 1, c1["number"].(map[string]interface{})["value"])
		assert.Equal(t, "FileSource", c1["number"].(map[string]interface{})["source"])
	})
}
func TestConfig_GetInt(t *testing.T) {
	s := archaius.GetInt("number", 0)
	assert.Equal(t, 1, s)
	s2 := archaius.GetInt64("number", 0)
	var a int64 = 1
	assert.Equal(t, a, s2)
}
func TestConfig_RegisterListener(t *testing.T) {
	eventHandler := EListener{}
	err := archaius.RegisterListener(eventHandler, "a*")
	assert.NoError(t, err)
	defer archaius.UnRegisterListener(eventHandler, "a*")

}

func TestUnmarshalConfig(t *testing.T) {
	b := []byte(`
key: peter
info:
  address: a
  number: 8
metadata_str:
  key01: "value01"
  key02: "value02"
metadata_int:
  key01: 1
  key02: 2
metadata_struct:
  key01: {address: "addr03",number: 1230}
  key02: {address: "addr04",number: 1231}
metadata_ptr:
  key01: {address: "addr05",number: 1232}
  key02: {address: "addr06",number: 1233}
str_arr:
  - "list01"
  - "list02"
int_arr:
  - 1
  - 2
infos:
  - address: "addr01"
    number: 100
    users:
      - name: "yourname"
        age: 21
infos_ptr:
  - address: "addr02"
    number: 123
    users:
      - name: "yourname1"
        age: 22
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f3.yaml")
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	err = archaius.Init(archaius.WithMemorySource())
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(filename1)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	type User struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}

	type Info struct {
		Addr   string `yaml:"address"`
		Number int    `yaml:"number"`
		Us     []User `yaml:"users"`
	}
	type Person struct {
		Name     string            `yaml:"key"`
		MDS      map[string]string `yaml:"metadata_str"`
		MDI      map[string]int    `yaml:"metadata_int"`
		MDSTR    map[string]Info   `yaml:"metadata_struct"`
		MDPTR    map[string]*Info  `yaml:"metadata_ptr"`
		Info     *Info             `yaml:"info"`
		StrArr   []string          `yaml:"str_arr"`
		IntArr   []int             `yaml:"int_arr"`
		Infos    []Info            `yaml:"infos"`
		InfosPtr []*Info           `yaml:"infos_ptr"`
	}
	err = archaius.AddFile(filename1)
	assert.NoError(t, err)
	p := &Person{}
	err = archaius.UnmarshalConfig(p)
	assert.NoError(t, err)
	assert.Equal(t, "peter", p.Name)
	// case map[string]string
	assert.Equal(t, "value01", p.MDS["key01"])
	// case map[string]int
	assert.Equal(t, 1, p.MDI["key01"])
	// case map[string]struct
	assert.Equal(t, "addr03", p.MDSTR["key01"].Addr)
	// case map[string]ptr
	assert.Equal(t, "addr05", p.MDPTR["key01"].Addr)

	// case ptr
	assert.Equal(t, "a", p.Info.Addr)
	// case ptr
	assert.Equal(t, 8, p.Info.Number)
	// case string array
	assert.Equal(t, "list01", p.StrArr[0])
	// case int array
	assert.Equal(t, 1, p.IntArr[0])
	// case struct array
	assert.Equal(t, "addr01", p.Infos[0].Addr)
	// case struct array
	assert.Equal(t, "yourname", p.Infos[0].Us[0].Name)
	// case ptr array
	assert.Equal(t, "addr02", p.InfosPtr[0].Addr)
	assert.Equal(t, "yourname1", p.InfosPtr[0].Us[0].Name)
	t.Run("map test", func(t *testing.T) {
		archaius.Set("key", "peter")
		archaius.Set("metadata.name", "123")
		type Person struct {
			Name string            `yaml:"key"`
			MD   map[string]string `yaml:"metadata"`
		}
		p := &Person{}
		err := archaius.UnmarshalConfig(p)
		assert.NoError(t, err)
		assert.Equal(t, "peter", p.Name)
		assert.Equal(t, "123", p.MD["name"])
	})
}

func TestMarshalConfig(t *testing.T) {
	b := []byte(`
info:
  address: a
metadata_str:
  key01: "value01"
metadata_int:
  key01: 1
metadata_struct:
  key01: {address: "addr03",number: 1230}
metadata_ptr:
  key01: {address: "addr05",number: 1232}
str_arr:
  - "list01"
int_arr:
  - 1
infos:
  - address: "addr01"
    users:
      - name: "yourname"
infos_ptr:
  - number: 123
    users:
      - name: "yourname1"
        age: 22
`)
	d, _ := os.Getwd()
	filename1 := filepath.Join(d, "f4.yaml")
	f1, err := os.Create(filename1)
	assert.NoError(t, err)
	err = archaius.Init(archaius.WithMemorySource())
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(filename1)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	err = archaius.AddFile(filename1)
	assert.NoError(t, err)
	buf := bytes.NewBuffer(nil)
	err = archaius.WriteTo(buf)
	assert.NoError(t, err)
	t.Logf("%s", buf.String())
}

func TestInitConfigCenter(t *testing.T) {
	err := archaius.EnableRemoteSource("fake", nil)
	assert.Error(t, err)
}
func TestClean(t *testing.T) {
	err := archaius.Clean()
	assert.NoError(t, err)
	s := archaius.Get("age")
	assert.Equal(t, nil, s)
}
func TestInitConfigBool2String(t *testing.T) {
	b := []byte(`
ssl:
  rest.Provider.cipherPlugin: aes
  rest.Provider.verifyPeer: false
  rest.Provider.cipherSuits: TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
  rest.Provider.protocol: TLSv1.2
  rest.Provider.caFile: ca.cer
  rest.Provider.certFile: cert_chain.pem
  rest.Provider.keyFile: private.pem
  rest.Provider.certPwdFile: PwdFile.yaml
`)
	d, _ := os.Getwd()
	testFile := filepath.Join(d, "tls_bool.yaml")
	f1, err := os.Create(testFile)
	assert.NoError(t, err)
	err = archaius.Init(archaius.WithMemorySource())
	assert.NoError(t, err)
	defer f1.Close()
	defer os.Remove(testFile)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	type Ssl struct {
		Ssl map[string]string `yaml:"ssl"`
	}
	err = archaius.AddFile(testFile)
	assert.NoError(t, err)
	//p := &Person{}
	sslConfig := &Ssl{}
	err = archaius.UnmarshalConfig(sslConfig)
	assert.NoError(t, err)
	assert.Equal(t, "aes", sslConfig.Ssl["rest.Provider.cipherPlugin"])
	assert.Equal(t, "false", sslConfig.Ssl["rest.Provider.verifyPeer"])
	assert.Equal(t, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", sslConfig.Ssl["rest.Provider.cipherSuits"])
	assert.Equal(t, "TLSv1.2", sslConfig.Ssl["rest.Provider.protocol"])
	assert.Equal(t, "ca.cer", sslConfig.Ssl["rest.Provider.caFile"])
	assert.Equal(t, "cert_chain.pem", sslConfig.Ssl["rest.Provider.certFile"])
	assert.Equal(t, "private.pem", sslConfig.Ssl["rest.Provider.keyFile"])
	assert.Equal(t, "PwdFile.yaml", sslConfig.Ssl["rest.Provider.certPwdFile"])
}
