package configcenter_test

import (
	"testing"

	"github.com/go-chassis/go-archaius/pkg/configcenter"
)

func TestNew(t *testing.T) {
	configcenter.New(configcenter.Options{})
}
