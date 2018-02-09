package converter_test

import (
	"testing"

	"github.com/ServiceComb/go-archaius/converter"
	"github.com/ServiceComb/go-chassis/core/lager"
	"github.com/stretchr/testify/assert"
)

func TestConverter(t *testing.T) {
	data := []byte(`{"name":"yaml", "work":"marshal"}`)
	yamlContent, err := converter.Converter(data, "yaml")
	assert.NotEqual(t, yamlContent, nil)
	assert.Nil(t, err)
}

func TestConverterInvalid(t *testing.T) {
	lager.Initialize("", "INFO", "", "size", true, 1, 10, 7)
	data := []byte(`{"name":"yaml", "work":"marshal"}`)
	yamlContent, err := converter.Converter(data, "json")
	assert.Nil(t, yamlContent)
	assert.NotNil(t, err)
}

func TestConverterErrorScenario(t *testing.T) {
	data := []byte(`{"name":"yaml""", "work":"marshal"}`)
	yamlContent, err := converter.Converter(data, "yaml")
	assert.NotNil(t, err)
	assert.Nil(t, yamlContent)
}
