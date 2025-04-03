package kie

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-chassis/go-archaius/source/remote"
)

func TestGenerateLabels(t *testing.T) {
	optionsLabels := map[string]string{
		remote.LabelApp:         "app",
		remote.LabelEnvironment: "env",
		remote.LabelService:     "service",
		remote.LabelVersion:     "1.0.0",
		"foo":                   "bar",
	}
	dimensionApp, err := GenerateLabels(DimensionApp, optionsLabels)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		remote.LabelApp:         "app",
		remote.LabelEnvironment: "env",
	}, dimensionApp)

	dimensionService, err := GenerateLabels(DimensionService, optionsLabels)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		remote.LabelApp:         "app",
		remote.LabelEnvironment: "env",
		remote.LabelService:     "service",
	}, dimensionService)
}
