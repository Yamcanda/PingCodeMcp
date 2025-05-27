package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
)

// DoGet 发送 GET 请求，支持自定义 header 和 query 参数
func DoGet(ctx context.Context, rawURL string, headers map[string]string, query map[string]string) ([]byte, *http.Response, error) {
	// 构建 URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, err
	}
	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return body, resp, nil
}

// DoPost 发送 POST 请求，支持自定义 header、body
func DoPost(ctx context.Context, rawURL string, headers map[string]string, body io.Reader) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return respBody, resp, nil
}

// DoPostJSON 发送 JSON 格式的 POST 请求
func DoPostJSON(ctx context.Context, rawURL string, headers map[string]string, jsonBody []byte) ([]byte, *http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"
	return DoPost(ctx, rawURL, headers, bytes.NewReader(jsonBody))
}

// DoPatch 发送 PATCH 请求，支持自定义 header、body
func DoPatch(ctx context.Context, rawURL string, headers map[string]string, body []byte) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return respBody, resp, nil
}
