package maputil_test

import (
	"github.com/go-chassis/go-archaius/pkg/maputil"
	"testing"
)

func TestMap2String(t *testing.T) {
	m := make(map[string]string)
	m["s"] = "a"
	m["c"] = "c"
	m["d"] = "b"
	t.Log(maputil.Map2String(m))
}
