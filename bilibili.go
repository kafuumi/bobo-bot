package main

import (
	"bytes"
	"errors"
	"github.com/Hami-Lemon/bobo-bot/logger"
	"github.com/Hami-Lemon/bobo-bot/request"
	"github.com/Hami-Lemon/bobo-bot/util"
	"github.com/tidwall/gjson"
	"strconv"
)

//用于身份授权的 cookie 的键名
const (
	DedeUserID    = "DedeUserID"
	DedeUserIDMd5 = "DedeUserID__ckMd5"
	SessData      = "SESSDATA"
	Csrf          = "bili_jct"
	SId           = "sid"
)

// Account 普通用户
type Account struct {
	uname string //该账号的昵称
	uid   uint64 //该账号的uid
	alias string //别名
}

// MonitorAccount 赛博监控账号
type MonitorAccount struct {
	Account
	follower int    //粉丝数
	face     string //头像
	sign     string //签名
}

// BotAccount bot所登录的账号
type BotAccount struct {
	Account
	uidMd5   string //cookies中的DedeUserID__ckMd5
	sessData string //cookies中的SESSDATA
	csrf     string //cookies中的bili_jct，部分接口请求参数中的csrf也是该值
	sid      string //cookies中的sid
}

// Comment 一条评论
type Comment struct {
	Account
	ctime    uint64 //评论发布的时间戳，单位秒
	msg      string //评论内容
	replyId  uint64 //评论id
	typeCode int    //评论区类型码
	oid      uint64 //评论区的id
}

// Board 评论区，或者叫版聊区
type Board struct {
	name     string //该评论区名称
	dId      uint64 //评论区所在动态的id
	oid      uint64 //该评论区的id
	typeCode int    //该评论区的类型码
	allCount int    //总评论数,包含楼中楼
	count    int    //评论数，不包含楼中楼
}

// BiliBili 与b站后台接口交互的对象
type BiliBili struct {
	user   BotAccount
	client *request.Client
	logger *logger.Logger //日志
}

func checkResp(entity request.Entity, err error) (*gjson.Result, error) {
	if util.IsError(err, "request fail!") {
		return nil, err
	}
	reader := entity.Reader().(*bytes.Buffer)
	//使用 gjson 库获取响应体中的数据
	result := gjson.ParseBytes(reader.Bytes())
	//code 不为0，出现错误
	if result.Get("code").Int() != 0 {
		msg := result.Get("message").String()
		return nil, errors.New(msg)
	}
	data := result.Get("data")
	//没有 data 字段
	if !data.Exists() {
		return nil, nil
	}
	return &data, nil
}

func BiliBiliLogin(user BotAccount) *BiliBili {
	header := map[string]string{
		"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36",
		"Accept-Language":    "zh-CN,zh;q=0.9",
		"Accept-Encoding":    "gzip, deflate, br",
		"sec-ch-ua":          `ot A;Brand";v="99", "Chromium";v="96", "Google Chrome";v="96`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "Windows",
	}

	cookie := map[string]string{
		DedeUserID:    strconv.FormatUint(user.uid, 10),
		DedeUserIDMd5: user.uidMd5,
		SessData:      user.sessData,
		Csrf:          user.csrf,
		SId:           user.sid,
	}
	biliLogger := logger.New("BiliBili", logLevel, logDst)
	client := request.New(header, cookie, 5)
	//获取用户名，判断该 cookie 是否有效
	urlStr := "https://api.bilibili.com/x/member/web/account"
	data, err := checkResp(client.Get(urlStr, nil, nil))
	if err != nil {
		biliLogger.Error("登录失败：%v", err)
		return nil
	}
	//用户名
	user.uname = data.Get("uname").String()
	user.alias = "bot"
	biliLogger.Debug("登录成功! uname: %s", user.uname)
	return &BiliBili{
		user:   user,
		client: client,
		logger: biliLogger,
	}
}

// LikeComment 点赞评论
func (b *BiliBili) LikeComment(comment Comment) bool {
	urlStr := "https://api.bilibili.com/x/v2/reply/action"
	body := request.NewNameValeEntity(
		map[string]interface{}{
			"type":     comment.typeCode,
			"oid":      comment.oid,
			"rpid":     comment.replyId,
			"action":   1,
			"csrf":     b.user.csrf,
			"ordering": "time",
		}, request.ApplicationUrlencoded)

	_, err := checkResp(b.client.Post(urlStr, nil, body))
	if err != nil {
		b.logger.Error("点赞评论失败：%v", err)
		pushAndLog(b.logger, "点赞评论失败：%v", err)
		return false
	}
	b.logger.Debug("成功点赞：%s uname: %s uid: %d",
		comment.msg, comment.uname, comment.uid)
	return true
}

// HateComment 点踩
func (b *BiliBili) HateComment() {
	//https://api.bilibili.com/x/v2/reply/hate
	//oid=197316850&type=11&rpid=117049115424&action=1&ordering=time&jsonp=jsonp&csrf=ec8e3
}

// ReportComment 举报评论
func (b *BiliBili) ReportComment() {
	//https://api.bilibili.com/x/v2/reply/report
	//oid=197316850&type=11&rpid=117049115424&reason=4&content=&ordering=time&jsonp=jsonp&csrf=ec8e
}

// GetCommentsPage 获取评论区的评论数
func (b *BiliBili) GetCommentsPage(board *Board) bool {
	urlStr := "https://api.bilibili.com/x/v2/reply/main"
	params := map[string]interface{}{
		"oid":  board.oid,
		"type": board.typeCode,
		"mode": 2, //按时间排序
		"ps":   1, //只获取一条评论
	}
	data, err := checkResp(b.client.GetWithRetry(urlStr, params, nil, 2))
	if err != nil {
		b.logger.Error("获取评论数量失败：oid: %d, %v", board.oid, err)
		pushAndLog(b.logger, "获取评论数量失败：oid: %d, %v", board.oid, err)
		return false
	}
	cursor := data.Get("cursor")
	board.allCount = int(cursor.Get("all_count").Int())
	board.count = int(cursor.Get("prev").Int())
	b.logger.Debug("获取评论数成功，all_count:%d, count:%d", board.allCount, board.count)
	return true
}

// GetComments 获取评论
func (b *BiliBili) GetComments(board Board) []Comment {
	urlStr := "https://api.bilibili.com/x/v2/reply/main"
	params := map[string]interface{}{
		"oid":  board.oid,
		"type": board.typeCode,
		"mode": 2, //按时间排序
	}

	data, err := checkResp(b.client.Get(urlStr, params, nil))
	if err != nil {
		b.logger.Error("获取评论失败：oid: %d, %v", board.oid, err)
		return nil
	}
	//获取评论，默认获取20条
	replies := data.Get("replies").Array()
	repliesLen := len(replies)
	comments := make([]Comment, repliesLen)
	for i := repliesLen - 1; i >= 0; i-- {
		reply := replies[i]
		comment := Comment{
			Account: Account{
				uid:   reply.Get("mid").Uint(),
				uname: reply.Get("member.uname").String(),
			},
			ctime:    reply.Get("ctime").Uint(),
			msg:      reply.Get("content.message").String(),
			replyId:  reply.Get("rpid").Uint(),
			typeCode: board.typeCode,
			oid:      board.oid,
		}
		comments[i] = comment
		b.logger.Debug("获取到评论：%#v", comment)
	}
	b.logger.Debug("获取评论成功：oid: %d, 获取评论数：%d", board.oid, len(comments))
	return comments
}

// PostComment 发评论，board 为对应的评论区，comment 不为空则表示评论区中回复对应的评论
func (b *BiliBili) PostComment(board Board, comment *Comment, msg string) bool {
	urlStr := "https://api.bilibili.com/x/v2/reply/add"
	body := request.NewNameValeEntity(map[string]interface{}{
		"type":    board.typeCode,
		"oid":     board.oid,
		"message": msg,
		"plat":    1,
		"csrf":    b.user.csrf,
	}, request.ApplicationUrlencoded)
	if comment != nil {
		body.Add("root", comment.replyId)
		body.Add("parent", comment.replyId)
		b.logger.Debug("发送楼中楼评论，对应楼：%s", comment.msg)
	}
	_, err := checkResp(b.client.Post(urlStr, nil, body))
	if err != nil {
		b.logger.Error("发布评论失败：oid: %d, msg: %s, err: %v",
			board.oid, msg, err)
		return false
	}
	b.logger.Debug("发布评论成功：oid: %d, msg: %s", board.oid, msg)
	return true
}

// BoardDetail 获取评论区详细信息
func (b *BiliBili) BoardDetail(board *Board) bool {
	urlStr := "https://api.bilibili.com/x/polymer/web-dynamic/v1/detail"
	params := map[string]interface{}{
		"timezone_offset": 0,
		"id":              board.dId,
	}
	data, err := checkResp(b.client.Get(urlStr, params, nil))
	if err != nil {
		b.logger.Error("获取评论区信息失败，oid: %d, err: %v", board.oid, err)
		return false
	}
	board.oid, _ = strconv.ParseUint(data.Get("item.basic.comment_id_str").String(),
		10, 64)
	board.typeCode = int(data.Get("item.basic.comment_type").Int())
	//board.allCount = int(data.Get("modules.module_stat.comment.count").Int())
	if board.name == "" {
		board.name = "未命名版"
	}
	b.logger.Debug("评论区信息：type: %d, allCount: %d", board.typeCode, board.allCount)
	return true
}

func (b *BiliBili) AccountSpace(account MonitorAccount) {
	//https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?offset=&host_mid=33605910&timezone_offset=-480
}

// AccountStat 获取账号粉丝数
func (b *BiliBili) AccountStat(account *MonitorAccount) bool {
	//https://api.bilibili.com/x/relation/stat?vmid=33605910&jsonp=jsonp
	urlStr := "https://api.bilibili.com/x/relation/stat"
	params := map[string]interface{}{
		"vmid": account.uid,
	}
	data, err := checkResp(b.client.Get(urlStr, params, nil))
	if err != nil {
		b.logger.Error("获取粉丝数失败：uid：%d, err: %v", account.uid, err)
		return false
	}
	account.follower = int(data.Get("follower").Int())
	b.logger.Debug("获取粉丝数：uid: %d, follower: %d", account.uid, account.follower)
	return true
}

// AccountInfo 获取详细信息：用户昵称，头像，签名
func (b *BiliBili) AccountInfo(account *MonitorAccount) bool {
	urlStr := "https://api.bilibili.com/x/space/acc/info"
	params := map[string]interface{}{
		"mid": account.uid,
	}
	data, err := checkResp(b.client.Get(urlStr, params, nil))
	if err != nil {
		b.logger.Error("获取用户信息失败：uid: %d, err: %v", account.uid, err)
		return false
	}
	//用户名
	account.uname = data.Get("name").String()
	if account.alias == "" {
		account.alias = account.uname
	}
	//头像
	account.face = data.Get("face").String()
	//签名
	account.sign = data.Get("sign").String()
	b.logger.Debug("获取用户信息：uid: %d, uname: %s, alias: %s, face: %s, sign: %s",
		account.uid, account.uname, account.alias, account.face, account.sign)
	return true
}
