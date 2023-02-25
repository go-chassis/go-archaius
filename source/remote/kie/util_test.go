package kie

import (
	"testing"

	"github.com/arielsrv/go-archaius/source/remote"
	"github.com/stretchr/testify/assert"
)

func TestGenerateLabels(t *testing.T) {
	optionsLabels := map[string]string{
		remote.LabelApp:         "app",
		remote.LabelEnvironment: "env",
		remote.LabelService:     "service",
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
