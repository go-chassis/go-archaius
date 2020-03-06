package configcenter_test

import (
	"testing"

	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-archaius/source/remote/configcenter"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigCenterSource(t *testing.T) {
	opts := &archaius.RemoteInfo{
		DefaultDimension: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		URL:        "http://",
	}
	_, err := configcenter.NewConfigCenterSource(opts)
	assert.NoError(t, err)
}
