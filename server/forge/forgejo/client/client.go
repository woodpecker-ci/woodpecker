// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

var jsonHeader = http.Header{"content-type": []string{"application/json"}}

type Forgejo struct {
	url         string
	accessToken string
	client      *http.Client
	ctx         context.Context
	logger      zerolog.Logger
}

type Response struct {
	*http.Response
	data []byte
}

func NewClient(ctx context.Context, logger zerolog.Logger, url, accessToken string, client *http.Client) (*Forgejo, error) {
	forgejo := &Forgejo{
		url:         strings.TrimSuffix(url, "/"),
		accessToken: accessToken,
		client:      client,
		ctx:         ctx,
		logger:      logger,
	}

	return forgejo, nil
}

type ListOptions struct {
	// Setting Page to -1 disables pagination on endpoints that support it.
	// Page numbering starts at 1.
	Page int
	// The default value depends on the server config DEFAULT_PAGING_NUM
	// The highest valid value depends on the server config MAX_RESPONSE_ITEMS
	PageSize int
}

func (o ListOptions) getURLQuery() url.Values {
	query := make(url.Values)
	query.Add("page", fmt.Sprintf("%d", o.Page))
	query.Add("limit", fmt.Sprintf("%d", o.PageSize))

	return query
}

// setDefaults applies default pagination options.
// If .Page is set to -1, it will disable pagination.
// WARNING: This function is not idempotent, make sure to never call this method twice!
func (o *ListOptions) setDefaults() {
	if o.Page < 0 {
		o.Page, o.PageSize = 0, 0
		return
	} else if o.Page == 0 {
		o.Page = 1
	}
}

func (f *Forgejo) StatusCode(resp *Response) (int, []byte, error) {
	if resp.StatusCode/100 == 2 {
		return resp.StatusCode, nil, nil
	}
	data, err := f.readData(resp)
	if err != nil {
		return -1, nil, err
	}
	return resp.StatusCode, data, nil
}

type ForgejoError struct {
	Status  int
	Message string
}

func (e ForgejoError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, e.Message)
}

func (f *Forgejo) StatusCodeToError(resp *Response) error {
	status, message, err := f.StatusCode(resp)
	if err != nil {
		return err
	}
	if status/100 == 2 {
		return nil
	}
	return ForgejoError{
		Status:  status,
		Message: string(message),
	}
}

func (f *Forgejo) getStatusCode(method, path string, header http.Header, body io.Reader) (int, *Response, error) {
	resp, err := f.doRequest(method, path, header, body)
	if err != nil {
		return -1, nil, err
	}
	return resp.StatusCode, resp, err
}

func (f *Forgejo) getParsedResponse(method, path string, header http.Header, body io.Reader, obj interface{}) (*Response, error) {
	data, resp, err := f.getResponse(method, path, header, body)
	if err != nil {
		return resp, err
	}
	if err := f.StatusCodeToError(resp); err != nil {
		return resp, err
	}
	return resp, json.Unmarshal([]byte(data), obj)
}

func (f *Forgejo) getResponse(method, path string, header http.Header, body io.Reader) ([]byte, *Response, error) {
	resp, err := f.doRequest(method, path, header, body)
	if err != nil {
		return nil, resp, err
	}
	data, err := f.readData(resp)
	if err != nil {
		return nil, resp, err
	}
	return data, resp, nil
}

func (f *Forgejo) readData(resp *Response) ([]byte, error) {
	if resp.data == nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("Body read on HTTP error %d: %v", resp.StatusCode, err)
		}
		f.logger.Debug().Msgf("Response Body: %s\n", data)
		resp.data = data
	}
	return resp.data, nil
}

func (f *Forgejo) doRequest(method, path string, header http.Header, body io.Reader) (*Response, error) {
	if f.logger.GetLevel() <= zerolog.DebugLevel {
		var bodyStr string
		if body != nil {
			bs, _ := io.ReadAll(body)
			body = bytes.NewReader(bs)
			bodyStr = string(bs)
		}
		f.logger.Debug().Msgf("%s: %s\nHeader: %v\nBody: %s\n", method, f.url+"/api/v1"+path, header, bodyStr)
	}
	req, err := http.NewRequestWithContext(f.ctx, method, f.url+"/api/v1"+path, body)
	if err != nil {
		return nil, err
	}
	if len(f.accessToken) != 0 {
		req.Header.Set("Authorization", "token "+f.accessToken)
	}

	for k, v := range header {
		req.Header[k] = v
	}

	resp, err := f.client.Do(req)
	f.logger.Debug().Msgf("Response: %v %v\n", err, resp)
	if err != nil {
		return nil, err
	}
	return &Response{resp, nil}, nil
}

func escapeValidatePathSegments(seg ...*string) error {
	for i := range seg {
		if seg[i] == nil || len(*seg[i]) == 0 {
			return fmt.Errorf("path segment [%d] is empty", i)
		}
		*seg[i] = url.PathEscape(*seg[i])
	}
	return nil
}

func pathEscapeSegments(path string) string {
	slice := strings.Split(path, "/")
	for index := range slice {
		slice[index] = url.PathEscape(slice[index])
	}
	escapedPath := strings.Join(slice, "/")
	return escapedPath
}
