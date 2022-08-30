package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

	os.Unsetenv("STR_ENV")
	str11 := "env:${STR_ENV^^||local}"
	str12 := "env:${STR_ENV,,||local}"
	str13 := "env:${STR_ENV^||local}"
	str14 := "env:${STR_ENV,||local}"
	assert.Equal(t, "env:local", ExpandValueEnv(str11))
	assert.Equal(t, "env:local", ExpandValueEnv(str12))
	assert.Equal(t, "env:local", ExpandValueEnv(str13))
	assert.Equal(t, "env:local", ExpandValueEnv(str14))

	os.Setenv("STR_ENV", "TesT")
	assert.Equal(t, "env:TEST", ExpandValueEnv(str11))
	assert.Equal(t, "env:test", ExpandValueEnv(str12))
	assert.Equal(t, "env:TesT", ExpandValueEnv(str13))
	assert.Equal(t, "env:tesT", ExpandValueEnv(str14))

	os.Setenv("STR_ENV", "工Test")
	assert.Equal(t, "env:工TEST", ExpandValueEnv(str11))
	assert.Equal(t, "env:工test", ExpandValueEnv(str12))
	assert.Equal(t, "env:工Test", ExpandValueEnv(str13))
	assert.Equal(t, "env:工Test", ExpandValueEnv(str14))

	os.Unsetenv("STR_ENV")
}
