package request

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	// ErrRequest 网络请求失败
	ErrRequest = errors.New("request fail")
)

type Client struct {
	header map[string]string
	cookie map[string]string
	client *http.Client
}

// New 根据指定的 header，cookie 和超时时间 timeout 创建一个 Client
//使用该 Client 发送的网络请求都会使用这里指定的 header 和 cookie
func New(header map[string]string, cookie map[string]string, timeout int) *Client {
	c := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	return &Client{
		header: header,
		cookie: cookie,
		client: c,
	}
}

//处理响应体数据
func handleResp(resp *http.Response) (Entity, error) {
	//获取响应体长度
	contentLength := resp.ContentLength
	//获取响应体数据类型
	contentType := resp.Header.Get("Content-Type")
	//获取压缩格式
	contentEncoding := resp.Header.Get("Content-Encoding")
	if contentLength < 0 {
		//响应头中不包含 content-length, 默认设为 1024，用于初始化 buffer
		contentLength = 1024
	}
	buf := bytes.NewBuffer(make([]byte, 0, contentLength))
	var src io.Reader = resp.Body
	//如果是gzip压缩，进行解压
	// TODO 其它压缩方式的解压处理
	if contentEncoding == "gzip" {
		var err error
		src, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
	}
	_, err := io.Copy(buf, src)
	if err != nil {
		return nil, err
	}
	return &ByteEntity{
		contentType: contentType,
		reader:      buf,
	}, nil
}

//发送网络请求
//urlStr 为请求地址；params 为 url 参数，可以为nil；body 为请求体,可以为 nil
func (c *Client) request(method, urlStr string,
	params map[string]interface{}, body Entity) (Entity, error) {
	//解析url参数
	v := url.Values{}
	for name, value := range params {
		v.Add(name, fmt.Sprintf("%v", value))
	}
	//创建请求
	var reader io.Reader
	if body != nil {
		reader = body.Reader()
	}
	req, err := http.NewRequest(method,
		fmt.Sprintf("%s?%s", urlStr, v.Encode()), reader)
	if err != nil {
		return nil, err
	}
	//设置cookie
	u := req.URL
	for name, value := range c.cookie {
		cookie := &http.Cookie{
			Name:   name,
			Value:  value,
			Path:   "/",
			Domain: u.Host,
			MaxAge: 0,
		}
		req.AddCookie(cookie)
	}
	//设置header
	for name, value := range c.header {
		req.Header.Add(name, value)
	}
	//发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return nil, ErrRequest
	}
	//如果响应头中带有 cookie，更新现有的 cookie
	for _, cookie := range resp.Cookies() {
		c.cookie[cookie.Name] = cookie.Value
	}
	//获取响应体的数据
	return handleResp(resp)
}

// Get 发送 GET 请求
func (c *Client) Get(urlStr string, params map[string]interface{}, body Entity) (Entity, error) {
	return c.request("GET", urlStr, params, body)
}

// Post 发送 POST 请求
func (c *Client) Post(urlStr string, params map[string]interface{}, body Entity) (Entity, error) {
	return c.request("POST", urlStr, params, body)
}
