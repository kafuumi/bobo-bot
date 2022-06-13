package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
)

const (
	ApplicationJson       = "application/json; charset=utf-8"
	ApplicationUrlencoded = "application/x-www-form-urlencoded"
)

// Entity 数据体接口， 表示请求体或响应体
type Entity interface {
	Reader() io.Reader
	ContentType() string
}

// ByteEntity 原始的二进制数据
type ByteEntity struct {
	contentType string    //数据类型
	reader      io.Reader //读取数据的 reader
}

func NewByteEntity(data []byte, contentType string) *ByteEntity {
	return &ByteEntity{
		reader:      bytes.NewReader(data),
		contentType: contentType,
	}
}

func (b *ByteEntity) ContentType() string {
	return b.contentType
}

func (b *ByteEntity) Reader() io.Reader {
	return b.reader
}

// NameValueEntity 键值对的数据体
type NameValueEntity struct {
	items       map[string]interface{}
	contentType string
}

func NewNameValeEntity(items map[string]interface{}, contentType string) *NameValueEntity {
	return &NameValueEntity{
		items:       items,
		contentType: contentType,
	}
}

func (n *NameValueEntity) ContentType() string {
	return n.contentType
}

func (n *NameValueEntity) Reader() io.Reader {
	buf := &bytes.Buffer{}
	switch n.contentType {
	case ApplicationJson:
		data, err := json.Marshal(n.items)
		if err != nil {
			return nil
		}
		buf.Write(data)
	case ApplicationUrlencoded:
		v := url.Values{}
		for name, value := range n.items {
			v.Add(name, fmt.Sprintf("%v", value))
		}
		buf.WriteString(v.Encode())
	}
	return buf
}

func (n *NameValueEntity) Add(name string, value interface{}) {
	n.items[name] = value
}
