package sources

import (
	"io"
	"os"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func Test_NewYamlConfirgurationSource(t *testing.T) {
	yamlContent := "a:\n  b:\n    c: valueC\n    d: valueD"

	filename := "/tmp/go-archaius-yamlconfigurationsource-test.yaml"
	var f *os.File
	os.Remove(filename)
	f, err := os.Create(filename)
	check(err)
	_, err = io.WriteString(f, yamlContent) //写入文件(字符串)
	check(err)

	c, err := NewYamlConfigurationSource(filename)
	check(err)
	if c.GetConfiguration()["a.b.c"] != "valueC" {
		t.Error("Got wrong value from yaml!!")
	}
}
