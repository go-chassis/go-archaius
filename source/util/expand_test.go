package util

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExpandValueEnv(t *testing.T) {
	str1 := "${NAME||archaius}=${IP||}:${PORT||8080}+++${0PORT||8080}+++${_PORT||8080}"
	assert.Equal(t, "archaius=:8080+++${0PORT||8080}+++8080", ExpandValueEnv(str1))

	str2 := "${IP||}:${PORT||8080}:${||8080}"
	assert.Equal(t, ":8080:${||8080}", ExpandValueEnv(str2))

	str3 := "${IP|}:${PORT||8080}"
	assert.Equal(t, "${IP|}:8080", ExpandValueEnv(str3))

	str4 := "        $${IP_09090||}  "
	assert.Equal(t, "$", ExpandValueEnv(str4))

	str5 := "${IP||}:8080${ADD R||:8080}"
	assert.Equal(t, ":8080${ADD R||:8080}", ExpandValueEnv(str5))

	str6 := "${IP||0.0.0.0}:8080"
	assert.Equal(t, "0.0.0.0:8080", ExpandValueEnv(str6))

	str7 := "${NAME||archaius}:${IP||github.com}"
	if e := os.Setenv("NAME", "go-archaius"); e != nil {
		t.Logf("err:%+v", e)
	}
	assert.Equal(t, "go-archaius:github.com", ExpandValueEnv(str7))

	str8 := "addr:${IP||127.0.0.1}:${PORT||8080}"
	if e := os.Setenv("IP", "0.0.0.0"); e != nil {
		t.Logf("err:%+v", e)
	}
	assert.Equal(t, "addr:0.0.0.0:8080", ExpandValueEnv(str8))

	str9 := "${POD_IP||}"
	if e := os.Setenv("POD_IP", "0.0.0.0"); e != nil {
		t.Logf("err:%+v", e)
	}
	assert.Equal(t, "0.0.0.0", ExpandValueEnv(str9))

	str10 := "${IP|}"
	if e := os.Setenv("IP", "0.0.0.0"); e != nil {
		t.Logf("err:%+v", e)
	}
	assert.Equal(t, "${IP|}", ExpandValueEnv(str10))
}
