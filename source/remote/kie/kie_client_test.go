package kie

import (
	"strconv"
	"testing"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/go-archaius/source/remote"
	"github.com/stretchr/testify/assert"
)

func TestNewKie(t *testing.T) {
	k, err := NewKie(remote.Options{
		ServerURI: "http://",
		Labels:    map[string]string{"app": "default"}})
	assert.NoError(t, err)
	assert.Equal(t, "default", k.Options().Labels["app"])
}

func TestMergeConfig(t *testing.T) {
	k, _ := NewKie(remote.Options{
		ServerURI: "http://",
		Labels: map[string]string{
			remote.LabelApp:         "app",
			remote.LabelEnvironment: "env",
			remote.LabelService:     "service",
			remote.LabelVersion:     "1.0.0",
		}})
	for i, dimension := range dimensionPrecedence {
		k.setDimensionConfigs(&model.KVResponse{
			Data: []*model.KVDoc{
				{
					Key:   "foo",
					Value: strconv.Itoa(i + 1),
				},
			},
		}, dimension)
	}
	assert.Equal(t, strconv.Itoa(len(dimensionPrecedence)), k.mergeConfig()["foo"].(string))
}
