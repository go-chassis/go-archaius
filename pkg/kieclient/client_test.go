/*
 * Copyright 2020 Huawei Technologies Co., Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kieclient

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Put(t *testing.T) {
	os.Setenv("HTTP_DEBUG", "1")
	c, _ := New(Config{
		Endpoint: "http://127.0.0.1:30110",
	})
	kv := KVRequest{
		Key:       "app.properties",
		Labels:    map[string]string{"service": "client"},
		Value:     "timeout: 1s",
		ValueType: "text",
	}
	result, err := c.Create(context.TODO(), kv, WithProject("client_test"))
	assert.NoError(t, err)
	kv.ID = result.ID
	_, err = c.Put(context.TODO(), kv, WithProject("client_test"))
	assert.NoError(t, err)
	kvs, responseRevision, _ := c.List(context.TODO(),
		WithKey("app.properties"),
		WithGetProject("client_test"),
		WithLabels(map[string]string{"service": "client"}))
	assert.GreaterOrEqual(t, len(kvs.Data), 1)

	_, _, err = c.List(context.TODO(),
		WithGetProject("client_test"),
		WithLabels(map[string]string{"service": "client"}),
		WithRevision(responseRevision))
	assert.Equal(t, ErrNoChanges, err)

	_, _, err = c.List(context.TODO(),
		WithGetProject("client_test"),
		WithLabels(map[string]string{"service": "client"}))
	assert.Error(t, err)

	_, _, err = c.List(context.TODO(),
		WithGetProject("client_test"),
		WithLabels(map[string]string{"service": "client"}),
		WithRevision(c.CurrentRevision()-1))
	assert.NoError(t, err)

	t.Run("long polling,wait 10s,change value,should return result", func(t *testing.T) {
		go func() {
			kvs, _, err = c.List(context.TODO(),
				WithLabels(map[string]string{"service": "client"}),
				WithGetProject("client_test"),
				WithWait("10s"))
			assert.NoError(t, err)
			assert.Equal(t, "timeout: 2s", kvs.Data[0].Value)
		}()
		kv := KVRequest{
			ID:     result.ID,
			Key:    "app.properties",
			Labels: map[string]string{"service": "client"},
			Value:  "timeout: 2s",
		}
		_, err := c.Put(context.TODO(), kv, WithProject("client_test"))
		assert.NoError(t, err)
	})
	t.Run("exact match", func(t *testing.T) {
		kv := KVRequest{
			Key:    "app.properties",
			Labels: map[string]string{"service": "client", "version": "1.0"},
			Value:  "timeout: 2s",
		}
		kv.ID = result.ID
		_, err := c.Put(context.TODO(), kv, WithProject("client_test"))
		assert.NoError(t, err)
		t.Log(c.CurrentRevision())
		kvs, _, err = c.List(context.TODO(),
			WithGetProject("client_test"),
			WithLabels(map[string]string{"service": "client"}),
			WithExact())
		assert.NoError(t, err)
		assert.Equal(t, 1, len(kvs.Data))
	})

}
func TestClient_Delete(t *testing.T) {
	c, err := New(Config{
		Endpoint: "http://127.0.0.1:30110",
	})

	kvBody := KVRequest{}
	kvBody.Key = "time"
	kvBody.Value = "100s"
	kvBody.ValueType = "text"
	kvBody.Labels = make(map[string]string)
	kvBody.Labels["env"] = "client_test"
	kv, err := c.Create(context.TODO(), kvBody, WithProject("client_test"))
	assert.NoError(t, err)
	kvs, _, err := c.List(context.TODO(),
		WithKey("time"),
		WithGetProject("client_test"),
		WithLabels(map[string]string{"env": "client_test"}))
	assert.NoError(t, err)
	assert.NotNil(t, kvs)
	err = c.Delete(context.TODO(), kv.ID, WithProject("client_test"))
	assert.NoError(t, err)
}
