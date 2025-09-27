package goweb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

// get网络请求
func Get(apiUrl string, params url.Values, res any) error {
	Url, err := url.Parse(apiUrl)
	if err != nil {
		return err
	}

	// 如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()

	// 发送请求
	resp, err := http.Get(Url.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// json解析
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	return nil
}

// post网络请求
func Post(apiUrl string, params url.Values, res any) error {
	resp, err := http.PostForm(apiUrl, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// json解析
	if err := json.Unmarshal(body, &res); err != nil {
		return err
	}
	return nil
}
