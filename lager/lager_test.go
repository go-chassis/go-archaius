package lager_test

import (
	"github.com/go-chassis/go-archaius/lager"
	paaslager "github.com/go-chassis/paas-lager"

	"testing"
)

func TestInitializewithNil(t *testing.T) {
	lager.InitLager(nil)
}

func TestInitializeWithValues(t *testing.T) {

	lager.InitLager(paaslager.NewLogger("log/archaius.log"))
}
