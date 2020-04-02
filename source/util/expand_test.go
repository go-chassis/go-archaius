package util

import (
	"os"
	"testing"
)

func TestExpandValueEnv(t *testing.T) {
	str := "${NAME||archaius}"
	if e := os.Setenv("NAM", "go-archaius"); e != nil {
		t.Logf("err:%+v", e)
	}
	realStr := ExpandValueEnv(str)
	t.Logf("realStr: %s", realStr)
}
