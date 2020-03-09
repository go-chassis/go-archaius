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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/foundation/security"
	"github.com/go-mesh/openlogging"
)

//match mode
const (
	QueryParamQ      = "q"
	QueryByLabelsCon = "&"
	QueryParamMatch  = "match"
	QueryParamKeyID  = "kv_id"
)

//http headers
const (
	HeaderRevision    = "X-Kie-Revision"
	HeaderContentType = "Content-Type"
)

//ContentType
const (
	ContentTypeJSON = "application/json"
)

//const
const (
	version   = "v1"
	APIPathKV = "kie/kv"

	schemeHTTPS  = "https"
	MsgGetFailed = "get failed"
	FmtGetFailed = "get %s failed,http status [%s], body [%s]"
)

//client errors
var (
	ErrKeyNotExist = errors.New("can not find value")
	ErrNoChanges   = errors.New("kv has not been changed since last polling")
)

//Client is the servicecomb kie rest client.
//it is concurrency safe
type Client struct {
	opts            Config
	cipher          security.Cipher
	c               *httpclient.Requests
	currentRevision int
}

//Config is the config of client
type Config struct {
	Endpoint      string
	DefaultLabels map[string]string
	VerifyPeer    bool //TODO make it works, now just keep it false
}

//New create a client
func New(config Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	httpOpts := &httpclient.Options{}
	if u.Scheme == schemeHTTPS {
		// #nosec
		httpOpts.TLSConfig = &tls.Config{
			InsecureSkipVerify: !config.VerifyPeer,
		}
	}
	c, err := httpclient.New(httpOpts)
	if err != nil {
		return nil, err
	}
	return &Client{
		opts: config,
		c:    c,
	}, nil
}

//Put create value of a key
func (c *Client) Put(ctx context.Context, kv KVRequest, opts ...OpOption) (*KVDoc, error) {
	options := OpOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.Project == "" {
		options.Project = defaultProject
	}
	url := fmt.Sprintf("%s/%s/%s/%s/%s", c.opts.Endpoint, version, options.Project, APIPathKV, kv.Key)
	h := http.Header{}
	h.Set(HeaderContentType, ContentTypeJSON)
	body, _ := json.Marshal(kv)
	resp, err := c.c.Do(ctx, http.MethodPut, url, h, body)
	if err != nil {
		return nil, err
	}
	b := ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		openlogging.Error(MsgGetFailed, openlogging.WithTags(openlogging.Tags{
			"k":      kv.Key,
			"status": resp.Status,
			"body":   b,
		}))
		return nil, fmt.Errorf(FmtGetFailed, kv.Key, resp.Status, b)
	}

	kvs := &KVDoc{}
	err = json.Unmarshal(b, kvs)
	if err != nil {
		openlogging.Error("unmarshal kv failed:" + err.Error())
		return nil, err
	}
	return kvs, nil
}

//Get get value of a key
func (c *Client) Get(ctx context.Context, opts ...GetOption) (*KVResponse, int, error) {
	options := GetOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.Project == "" {
		options.Project = defaultProject
	}
	if options.Revision == "" {
		options.Revision = strconv.Itoa(c.currentRevision)
	}
	var url string
	if options.Key != "" {
		url = fmt.Sprintf("%s/%s/%s/%s/%s?revision=%s", c.opts.Endpoint, version, options.Project, APIPathKV, options.Key, options.Revision)
	} else {
		url = fmt.Sprintf("%s/%s/%s/%s?revision=%s", c.opts.Endpoint, version, options.Project, APIPathKV, options.Revision)
	}
	if options.Wait != "" {
		url = url + "&wait=" + options.Wait
	}
	if options.Exact {
		url = url + "&" + QueryParamMatch + "=exact"
	}
	labels := ""
	if len(options.Labels) != 0 {
		for k, v := range options.Labels[0] {
			labels = labels + "&label=" + k + ":" + v
		}
		url = url + labels
	}
	h := http.Header{}
	resp, err := c.c.Do(ctx, http.MethodGet, url, h, nil)
	if err != nil {
		return nil, -1, err
	}
	responseRevision, err := strconv.Atoi(resp.Header.Get(HeaderRevision))
	if err != nil {
		responseRevision = -1
	}
	b := ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, responseRevision, ErrKeyNotExist
		}
		if resp.StatusCode == http.StatusNotModified {
			return nil, responseRevision, ErrNoChanges
		}
		openlogging.Error(MsgGetFailed, openlogging.WithTags(openlogging.Tags{
			"k":      options.Key,
			"status": resp.Status,
			"body":   b,
		}))
		return nil, responseRevision, fmt.Errorf(FmtGetFailed, options.Key, resp.Status, b)
	} else if err != nil {
		msg := fmt.Sprintf("get revision from response header failed when the request status is OK: %v", err)
		openlogging.Error(msg)
		return nil, responseRevision, fmt.Errorf(msg)
	}
	var kvs *KVResponse
	err = json.Unmarshal(b, &kvs)
	if err != nil {
		openlogging.Error("unmarshal kv failed:" + err.Error())
		return nil, responseRevision, err
	}
	c.currentRevision = responseRevision
	return kvs, responseRevision, nil
}

//Summary get value by labels
func (c *Client) Summary(ctx context.Context, opts ...GetOption) ([]*KVResponse, error) {
	options := GetOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.Project == "" {
		options.Project = defaultProject
	}
	labelParams := ""
	for _, labels := range options.Labels {
		labelParams += QueryParamQ + "="
		for labelKey, labelValue := range labels {
			labelParams += labelKey + ":" + labelValue + "+"
		}
		if labels != nil && len(labels) > 0 {
			labelParams = strings.TrimRight(labelParams, "+")
		}
		labelParams += QueryByLabelsCon
	}
	if options.Labels != nil && len(options.Labels) > 0 {
		labelParams = strings.TrimRight(labelParams, QueryByLabelsCon)
	}
	url := fmt.Sprintf("%s/%s/%s/%s?%s", c.opts.Endpoint, version, options.Project, "kie/summary", labelParams)
	h := http.Header{}
	resp, err := c.c.Do(ctx, http.MethodGet, url, h, nil)
	if err != nil {
		return nil, err
	}
	b := ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrKeyNotExist
		}
		openlogging.Error(MsgGetFailed, openlogging.WithTags(openlogging.Tags{
			"p":      options.Project,
			"status": resp.Status,
			"body":   b,
		}))
		return nil, fmt.Errorf("search %s failed,http status [%s], body [%s]", labelParams, resp.Status, b)
	}
	var kvs []*KVResponse
	err = json.Unmarshal(b, &kvs)
	if err != nil {
		openlogging.Error("unmarshal kv failed:" + err.Error())
		return nil, err
	}
	return kvs, nil
}

//Delete remove kv
func (c *Client) Delete(ctx context.Context, kvID, labelID string, opts ...OpOption) error {
	options := OpOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.Project == "" {
		options.Project = defaultProject
	}
	url := fmt.Sprintf("%s/%s/%s/%s/?%s=%s", c.opts.Endpoint, version, options.Project, APIPathKV,
		QueryParamKeyID, kvID)
	if labelID != "" {
		url = fmt.Sprintf("%s?labelID=%s", url, labelID)
	}
	h := http.Header{}
	h.Set(HeaderContentType, ContentTypeJSON)
	resp, err := c.c.Do(ctx, http.MethodDelete, url, h, nil)
	if err != nil {
		return err
	}
	b := ReadBody(resp)
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete %s failed,http status [%s], body [%s]", kvID, resp.Status, b)
	}
	return nil
}

//CurrentRevision return the current revision of kie, which is updated on the last get request
func (c *Client) CurrentRevision() int {
	return c.currentRevision
}

// ReadBody read body from the from the response
func ReadBody(resp *http.Response) []byte {
	if resp != nil && resp.Body != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			openlogging.Error(fmt.Sprintf("read body failed: %s", err.Error()))
			return nil
		}
		return body
	}
	openlogging.Error("response body or response is nil")
	return nil
}
