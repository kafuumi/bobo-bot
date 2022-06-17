package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Hami-Lemon/bobo-bot/logger"
	"github.com/Hami-Lemon/bobo-bot/set"
	"github.com/Hami-Lemon/bobo-bot/util"
)

const (
	CountCap = 24 * 60
)

type Counter struct {
	todayComment int            //统计时段内记录到的评论数
	peopleCount  map[uint64]int //参与评论的用户，记录不同用户的发评数量

	hotCount []int //统计时间段中，每一分钟内的评论数，数组索引表示距离统计开始时间的偏移量，单位分钟
	awlCount []int //每一分钟内的延迟统计

	maxCount   int //一分钟内的最多评论数
	maxDelay   int //最大延迟，单位秒
	maxComment int //单个账号发布的最多评论数

	startTime time.Time  //统计的开始时间点
	lock      sync.Mutex //互斥锁
}

// Reporter 延迟反馈报告
type Reporter struct {
	offset   int    //误差
	last     uint64 //上一次反馈时间
	interval int    // 两次反馈的间隔时间
}

type Bot struct {
	board   Board          //监控的评论区
	monitor MonitorAccount //监控的账户
	bili    *BiliBili
	counter *Counter //统计器
	logger  *logger.Logger
	isStop  bool //是否停止
	freshCD int  //抓取评论cd
	likeCD  int  //点赞cd
	report  *Reporter
}

func NewBot(bili *BiliBili, board Board,
	monitor MonitorAccount, freshCD, likeCD int) *Bot {
	bili.AccountInfo(&monitor)
	bili.AccountStat(&monitor)
	bili.BoardDetail(&board)
	now := time.Now()
	counter := Counter{
		peopleCount: make(map[uint64]int),
		hotCount:    make([]int, 0, CountCap),
		awlCount:    make([]int, 0, CountCap),
		startTime:   now,
	}

	return &Bot{
		board:   board,
		monitor: monitor,
		bili:    bili,
		counter: &counter,
		logger:  logger.New(fmt.Sprintf("Bot-%s", board.name), logLevel, logDst),
		isStop:  false,
		freshCD: freshCD,
		likeCD:  likeCD,
		report: &Reporter{
			offset:   freshCD,
			interval: 60, //一分钟只触发一次
		},
	}
}

// Monitor 开启赛博监控
func (b *Bot) Monitor() {
	tick := time.Tick(time.Duration(b.freshCD) * time.Second)
	//获取评论
	comments := b.bili.GetComments(b.board)
	if comments == nil {
		b.logger.Error("获取评论失败，oid=%d", b.board.oid)
		return
	}
	lastComments := set.NewSlice(comments)
	for range tick {
		if b.isStop {
			b.logger.Info("停止监控")
			break
		}
		comments = b.bili.GetComments(b.board)
		for _, comment := range comments {
			//该评论出现在上次获取到的评论中，可能已经点赞了
			if lastComments.Contains(comment) {
				continue
			}
			b.work(comment)
			b.counter.Count(comment)
			//TODO 监控个人资料修改 #3
			b.logger.Debug("点赞CD")
			time.Sleep(time.Duration(b.likeCD) * time.Second)
		}
		if comments == nil {
			b.logger.Error("获取评论失败，oid=%d, type=%d", b.board.oid, b.board.typeCode)
		} else {
			lastComments.Clear()
			lastComments.Add(comments...)
		}
		if b.isStop {
			b.logger.Info("停止监控")
			break
		}
		b.logger.Debug("刷新CD")
	}
}

func (b *Bot) work(comment Comment) {
	bili := b.bili
	//点赞该评论
	if bili.LikeComment(comment) {
		b.logger.Info("成功点赞评论, msg=%s, uname=%s, uid=%d",
			comment.msg, comment.uname, comment.uid)
	} else {
		b.logger.Error("点赞评论失败,oid=%d, rpid=%d, msg=%s",
			comment.oid, comment.replyId, comment.msg)
		return
	}
	//如果评论包含 test 触发延迟反馈
	if strings.Contains(comment.msg, "test") {
		//计算延迟，当前时间 - 评论发布时间 - freshCD
		//如果小于0，则延迟为0
		delay := b.report.Report(comment)
		if delay == "" {
			b.logger.Debug("间隔过短，不触发延迟反馈")
		} else {
			if bili.PostComment(b.board, &comment, delay) {
				b.logger.Info("反馈延迟：%s, rpid=%d, msg=%s", delay, comment.replyId, comment.msg)
			} else {
				b.logger.Error("反馈延迟失败")
			}
		}
	}
}

// Stop 停止赛博监控
func (b *Bot) Stop() {
	//TODO 修改结束策略 select 轮询方式 #4
	b.logger.Debug("调用停止函数")
	b.isStop = true
}

// Report 通过获取到评论的时间，减去评论的发出时间，计算延迟
func (r *Reporter) Report(comment Comment) string {
	now := uint64(time.Now().Unix())
	delay := int(now-comment.ctime) - r.offset
	// 因为设定每隔几秒获取一次评论，所以会存在几秒的误差，
	// 如果计算的延迟小于该间隔时间，则延迟为0
	if delay < 0 {
		delay = 0
	}
	if r.last == 0 || int(now-r.last) > r.interval {
		r.last = now
	} else {
		//间隔过短，不触发延迟反馈
		return ""
	}
	var delayMsg string
	if delay <= 60 {
		delayMsg = fmt.Sprintf("延迟为%2d秒", delay)
	} else if delay <= 60*60 {
		delayMsg = fmt.Sprintf("延迟为%d分%02d秒", delay/60, delay%60)
	} else {
		s := delay % 60
		delay /= 60
		m, h := delay%60, delay/60
		delayMsg = fmt.Sprintf("延迟为%d时%02d分%02d秒", h, m, s)
	}
	return delayMsg
}

// Count 评论数据计数
func (c *Counter) Count(comment Comment) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.peopleCount[comment.uid]++
	c.maxComment = util.MaxInt(c.maxComment, c.peopleCount[comment.uid])
	c.todayComment++

	ctime := int64(comment.ctime)
	index := int(ctime - c.startTime.Unix())
	if index >= 0 {
		index /= 60
		var hot int
		c.hotCount, hot = util.SliceGet(c.hotCount, index)
		hot++
		c.hotCount = util.SliceSet(c.hotCount, index, hot)
		c.maxCount = util.MaxInt(c.maxCount, c.hotCount[index])
	}

	now := time.Now().Unix()
	delay := int(now - ctime)
	index = int(now-c.startTime.Unix()) / 60
	var d int
	c.awlCount, d = util.SliceGet(c.awlCount, index)
	//只记录最大延迟时间，单位：秒
	if delay > d {
		c.awlCount = util.SliceSet(c.awlCount, index, delay)
		c.maxDelay = util.MaxInt(c.maxDelay, delay)
	}
}

//重置
func (c *Counter) reset() {
	//重置
	c.todayComment = 0
	c.peopleCount = make(map[uint64]int)
	c.hotCount = make([]int, 0, CountCap)
	c.awlCount = make([]int, 0, CountCap)
	c.maxCount = 0
	c.maxDelay = 0
	c.maxComment = 0
	c.startTime = time.Now()
}

// Summarize 总结评论数据
func (b *Bot) Summarize() {
	counter := b.counter
	counter.lock.Lock()
	defer counter.lock.Unlock()

	report := struct {
		Name       string `json:"name"`
		Oid        uint64 `json:"oid"`
		Start      int64  `json:"start"`
		Hot        []int  `json:"hot"`
		Awl        []int  `json:"awl"`
		People     int    `json:"people"`
		Count      int    `json:"count"`
		MaxCount   int    `json:"maxCount"`
		MaxDelay   int    `json:"maxDelay"`
		MaxComment int    `json:"maxComment"`
	}{
		Name:       b.board.name,
		Oid:        b.board.oid,
		Start:      counter.startTime.Unix(),
		Hot:        counter.hotCount,
		Awl:        counter.awlCount,
		People:     len(counter.peopleCount),
		Count:      counter.todayComment,
		MaxCount:   counter.maxCount,
		MaxDelay:   counter.maxDelay,
		MaxComment: counter.maxComment,
	}
	reportJson, _ := json.Marshal(report)
	now := time.Now()
	fileName := fmt.Sprintf("./report/%s.json", now.Format("200601021504"))
	jsonFile, err := os.Create(fileName)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir("./report", os.ModePerm)
		if util.IsError(err, "creat dir report fail!") {
			_, _ = io.Copy(os.Stdout, bytes.NewReader(reportJson))
			return
		}
		jsonFile, _ = os.Create(fileName)
	}
	_, _ = io.Copy(jsonFile, bytes.NewReader(reportJson))
	_ = jsonFile.Close()
	counter.reset()
}

// MonitorDynamic 动态监控 TODO
func (b *Bot) MonitorDynamic() {

}
