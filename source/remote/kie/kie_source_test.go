package kie

import (
	"testing"

	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
)

func TestNewKieSource(t *testing.T) {
	opts := &archaius.RemoteInfo{
		DefaultDimension: map[string]string{
			"app":         "default",
			"serviceName": "cart",
		},
		TenantName: "default",
		URL:        "http://",
		ClientType: "mock-client",
	}
	_, err := NewKieSource(opts)
	assert.NoError(t, err)
}
