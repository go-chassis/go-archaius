package configcenter_test

import (
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/go-chassis/go-archaius/source/remote/configcenter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewConfigCenter(t *testing.T) {
	c, err := configcenter.NewConfigCenter(remote.Options{
		ServerURI: "http://",
		Labels:    map[string]string{"app": "default"}})
	assert.NoError(t, err)
	assert.Equal(t, "default", c.Options().Labels["app"])
}
