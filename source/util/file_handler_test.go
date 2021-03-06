package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert2JavaProps(t *testing.T) {
	b := []byte(`
a: 1
b: 2
c:
 d: 3
e:
 - addr: "addvalue"
   nameber: ${NAME||10}
 - ${TEST||none}
`)
	os.Setenv("NAME", "go-archaius")
	m, err := Convert2JavaProps("test.yaml", b)
	assert.NoError(t, err)
	assert.Equal(t, m["c.d"], 3)

	e1 := m["e"]
	v, ok := e1.([]interface{})
	assert.True(t, ok)
	map1, ok1 := v[0].(map[string]interface{})
	assert.True(t, ok1)
	assert.Equal(t, "addvalue", map1["addr"])
	assert.Equal(t, "go-archaius", map1["nameber"])
	assert.Equal(t, "none", v[1])
}

func TestConvert2ConfigMap(t *testing.T) {
	b := []byte(`
a: 1
b: 2
c:
 d: 3
`)
	m, err := UseFileNameAsKeyContentAsValue("/root/test.yaml", b)
	assert.NoError(t, err)
	assert.Equal(t, b, m["test.yaml"])
}
