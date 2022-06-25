package push

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Pusher interface {
	Push(msg string, args ...any) error
}

// DingPusher 钉钉机器人消息推送
type DingPusher struct {
	webhook string //webhook地址
	secret  string //签名密钥
}

func NewDingPusher(webhook, secret string) *DingPusher {
	return &DingPusher{
		webhook: webhook,
		secret:  secret,
	}
}

func (d *DingPusher) Push(msg string, args ...any) error {
	//未配置webhook，不进行推送
	if strings.Compare("", d.webhook) == 0 {
		return nil
	}
	urlStr, err := d.sign()
	if err != nil {
		return err
	}
	text := fmt.Sprintf(msg, args...)
	body := map[string]interface{}{
		"at": map[string]interface{}{
			"isAtAll": true, //at全体成员
		},
		"text": map[string]interface{}{
			"content": text, //消息内容
		},
		"msgtype": "text", //消息为文本类型
	}
	jsonBody, _ := json.Marshal(body)
	reader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, urlStr, reader)
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second} //两秒的超时
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var errInfo struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err = json.NewDecoder(resp.Body).Decode(&errInfo)
	if err != nil {
		return err
	}
	if errInfo.ErrCode != 0 {
		return errors.New(errInfo.ErrMsg)
	}
	return nil
}

func (d *DingPusher) sign() (string, error) {
	if strings.Compare("", d.secret) == 0 {
		return d.webhook, nil
	}
	u, err := url.Parse(d.webhook)
	if err != nil {
		return "", err
	}
	//https://open.dingtalk.com/document/group/customize-robot-security-settings
	now := strconv.FormatInt(time.Now().UnixMilli(), 10)
	text := now + "\n" + d.secret
	h := hmac.New(sha256.New, []byte(d.secret))
	h.Write([]byte(text))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	query := u.Query()
	query.Add("sign", sign)
	query.Add("timestamp", now)
	u.RawQuery = query.Encode()
	return u.String(), nil
}
