package goweb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Http struct {
	Url     string
	Client  *http.Client
	Headers map[string]string
	Timeout time.Duration
	Error   error
}

// NewHttp 创建一个新的HTTP客户端实例
func NewHttp(url string) *Http {
	return &Http{
		Url:     url,
		Client:  &http.Client{},
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
}

// SetTimeout 设置请求超时时间
func (h *Http) SetTimeout(timeout time.Duration) *Http {
	h.Timeout = timeout
	return h
}

// SetHeader 设置请求头
func (h *Http) SetHeader(key, value string) *Http {
	h.Headers[key] = value
	return h
}

// Do 执行HTTP请求
func (h *Http) Do(method string, body io.Reader) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, h.Url, body)
	if err != nil {
		h.Error = err
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range h.Headers {
		req.Header.Set(key, value)
	}

	// 如果没有设置Content-Type，默认设置为application/x-www-form-urlencoded
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		h.Error = err
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("请求超时（%v）: %v", h.Timeout, err)
		}
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.Error = err
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.Error = fmt.Errorf("请求返回错误状态码: %d, 响应内容: %s", resp.StatusCode, string(respBody))
		return respBody, h.Error
	}

	return respBody, nil
}

// Get 发送GET请求
func Get(url string, params url.Values, result any) error {
	httpClient := NewHttp(url)

	// 如果有参数，添加到URL
	if len(params) > 0 {
		if strings.Contains(url, "?") {
			httpClient.Url += "&" + params.Encode()
		} else {
			httpClient.Url += "?" + params.Encode()
		}
	}

	data, err := httpClient.Do("GET", nil)
	if err != nil {
		return err
	}

	// 如果需要解析结果
	if result != nil {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("解析JSON失败: %v, 响应内容: %s", err, string(data))
		}
	}

	return nil
}

// PostJSON 发送JSON格式的POST请求
func PostJSON(url string, data any, result any) error {
	httpClient := NewHttp(url)
	httpClient.SetHeader("Content-Type", "application/json")

	// 将数据序列化为JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化JSON失败: %v", err)
	}

	responseData, err := httpClient.Do("POST", strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}

	// 如果需要解析结果
	if result != nil {
		if err := json.Unmarshal(responseData, result); err != nil {
			return fmt.Errorf("解析JSON失败: %v, 响应内容: %s", err, string(responseData))
		}
	}

	return nil
}
