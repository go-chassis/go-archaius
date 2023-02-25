package configcenter_test

import (
	"encoding/json"
	"testing"

	"github.com/arielsrv/go-archaius/pkg/configcenter"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigs(t *testing.T) {
	m := make(map[string]interface{})
	m["a"] = "b"
	m["c"] = "d"
	b, err := json.Marshal(m)
	assert.NoError(t, err)
	value := string(b)
	e := configcenter.Event{
		Action: "delete",
		Value:  value,
	}

	b, err = json.MarshalIndent(e, "", "  ")
	t.Log(string(b))
	assert.NoError(t, err)
	m2, err := configcenter.GetConfigs(b)
	assert.NoError(t, err)
	assert.Equal(t, "b", m2["a"])
}
