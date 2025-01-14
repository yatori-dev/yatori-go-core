package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// HttpClient 是一个可定制的 HTTP 客户端，用于发送请求
type HttpClient struct {
	client *http.Client
}

// NewHttpClient 创建一个新的 HttpClient 实例
func NewHttpClient() *HttpClient {
	return &HttpClient{client: &http.Client{}}
}

// SendRequest 执行一个带有可定制选项的 HTTP 请求
func (c *HttpClient) SendRequest(method string, url string, body interface{}, headers http.Header) ([]byte, http.Header, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, nil, errors.New("failed to marshal request body: " + err.Error())
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, nil, errors.New("failed to create HTTP request: " + err.Error())
	}
	switch {
	case headers == nil:
		// No headers provided
	case len(headers) == 0:
		// Empty headers provided (same as nil)
	default:
		req.Header = headers
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, errors.New("failed to execute HTTP request: " + err.Error())
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.New("failed to read response body: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New(fmt.Sprintf("HTTP request failed with status code %d: %s", resp.StatusCode, string(responseBody)))
	}
	return responseBody, resp.Header, nil
}
