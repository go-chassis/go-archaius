package remote_test

import (
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/go-chassis/go-archaius/source/remote/configcenter"
)

func TestEnable(t *testing.T) {
	_, err := remote.NewClient("config_center", remote.Options{
		ServerURI: "http://127.0.0.1:30100",
	})
	assert.Error(t, err)
}
