package goweb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Http struct {
	Url            string // 存储基础URL
	Client         *http.Client
	defaultHeaders map[string]string
	timeout        time.Duration
}

func NewHttp(url string) *Http {
	return &Http{
		Url:            url,
		Client:         &http.Client{},
		defaultHeaders: make(map[string]string),
		timeout:        30 * time.Second,
	}
}

// SetTimeout 设置超时时间
func (h *Http) SetTimeout(timeout time.Duration) *Http {
	h.timeout = timeout
	return h
}

// SetDefaultHeader 设置默认请求头
func (h *Http) SetHeader(key, value string) *Http {
	h.defaultHeaders[key] = value
	return h
}

// ClearDefaultHeaders 清除默认请求头
func (h *Http) ClearDefaultHeaders() *Http {
	h.defaultHeaders = make(map[string]string)
	return h
}

// Do 执行HTTP请求
func (h *Http) Do(urlPath string, method string, headers map[string]string, body io.Reader) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, h.Url+urlPath, body)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认请求头
	for key, value := range h.defaultHeaders {
		req.Header.Set(key, value)
	}

	// 设置本次请求特定的请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 如果请求体不为空且未设置Content-Type，设置默认值
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("请求超时（%v）: %w", h.timeout, err)
		}
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBody, fmt.Errorf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return respBody, nil
}

// Get 执行GET请求
func (h *Http) Get(params url.Values, result any) error {
	// 构建查询字符串
	queryString := ""
	if len(params) > 0 {
		queryString = "?" + params.Encode()
	}

	data, err := h.Do(queryString, "GET", nil, nil)
	if err != nil {
		return err
	}

	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析JSON失败: %w, 响应内容: %s", err, string(data))
		}
	}

	return nil
}

// PostForm 发送表单数据
func (h *Http) PostForm(formData url.Values, result any) error {
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	data, err := h.Do("", "POST", headers, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析JSON失败: %w, 响应内容: %s", err, string(data))
		}
	}

	return nil
}

// PostJSON 发送JSON数据
func (h *Http) PostJSON(jsonData any, result any) error {
	var body []byte
	var err error

	if jsonData != nil {
		body, err = json.Marshal(jsonData)
		if err != nil {
			return fmt.Errorf("序列化JSON失败: %w", err)
		}
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	data, err := h.Do("", "POST", headers, bytes.NewReader(body))
	if err != nil {
		return err
	}

	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析JSON失败: %w, 响应内容: %s", err, string(data))
		}
	}

	return nil
}

// PostMultipart 发送 multipart/form-data 请求
func (h *Http) PostMultipart(formBuilder func(w *multipart.Writer) error, result any) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 执行表单构建逻辑（由调用方实现）
	if err := formBuilder(writer); err != nil {
		return fmt.Errorf("构建multipart表单失败: %w", err)
	}

	// 关闭writer（生成最终的boundary）
	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭multipart writer失败: %w", err)
	}

	// 设置multipart请求头
	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}

	// 执行POST请求
	data, err := h.Do("", "POST", headers, body)
	if err != nil {
		return err
	}

	// 解析响应
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析JSON失败: %w, 响应内容: %s", err, string(data))
		}
	}

	return nil
}
