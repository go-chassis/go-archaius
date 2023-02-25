package event_test

import (
	"fmt"
	"testing"

	"github.com/arielsrv/go-archaius/event"
	"github.com/stretchr/testify/assert"
)

func TestPopulateEvents(t *testing.T) {
	events, err := event.PopulateEvents(
		"test",
		map[string]interface{}{
			"k1": "v1",
			"k3": "v2",
			"k4": "v4",
		},
		map[string]interface{}{
			"k1": "v1",
			"k2": "v2",
			"k3": "v3",
		})
	assert.NoError(t, err)
	for _, eve := range events {
		fmt.Printf("%+v\n", eve)
	}
	assert.Equal(t, 3, len(events))
}
