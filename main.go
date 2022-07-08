package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"github.com/Hami-Lemon/bobo-bot/logger"
	"github.com/Hami-Lemon/bobo-bot/push"
	"github.com/tidwall/gjson"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	Version     = "0.2.3"
	logFileSize = 1024 * 512
)

var (
	buildTime                 = "unknown time"
	logLevel                  = logger.Info
	logDst    logger.Appender = logger.NewFileAppender(logFileSize)
	//logDst     logger.Appender = logger.NewConsoleAppender()
	mainLogger  = logger.New("main", logLevel, logger.NewConsoleAppender())
	db          *DB
	pusher      push.Pusher //消息推送
	summaryFile = flag.String("r", "", "数据总结文件")
)

type config struct {
	BotOption
	isFans bool
	hour   int
	minute int
	dbname string
}

func main() {
	flag.Parse()
	mainLogger.Info("bobo-bot version: %s build on %s", Version, buildTime)
	botAccount, monitorAccount, board, con := readSetting()
	bili := BiliBiliLogin(botAccount)
	if bili == nil {
		mainLogger.Error("登录失败！")
		return
	} else {
		mainLogger.Info("登录成功，%s", bili.user.uname)
	}
	db = NewDB(con.dbname)
	if db == nil {
		return
	}
	var bot *Bot
	if strings.Compare("", *summaryFile) == 0 {
		bot = NewBot(bili, board, monitorAccount, con.BotOption)
	} else {
		mainLogger.Info("从上次中断中恢复...")
		f, err := os.Open(*summaryFile)
		if err != nil {
			mainLogger.Error("打开文件失败，%v", err)
			return
		}
		var summary Summary
		err = json.NewDecoder(f).Decode(&summary)
		if err != nil {
			mainLogger.Error("解析文件失败，%v", err)
			return
		}
		bot = RecoverBot(bili, con.BotOption, summary)
		mainLogger.Info("恢复信息：start=%s", bot.counter.startTime.Format("01-02 15:04:05"))
		mainLogger.Info("board:%d, allCount=%d, count=%d", bot.board.oid, bot.board.allCount, bot.board.count)
		mainLogger.Info("account:%d, uname=%s, follower=%d", bot.monitor.uid, bot.monitor.uname, bot.monitor.follower)
	}
	go waitExit(bot)
	go summarize(bot, con.hour, con.minute)
	go readCmd(bot)
	mainLogger.Info("开始赛博监控...")
	mainLogger.Info("监控评论区：%s, %d", board.name, board.dId)
	defer logDst.Close()
	if con.isFans {
		mainLogger.Info("粉丝数监控：uid=%d", monitorAccount.uid)
		go bot.MonitorFans()
	}
	bot.Monitor()
	bot.Summarize()
	db.Close()
	mainLogger.Info("程序停止")
}

func readCmd(bot *Bot) {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := sc.Text()
		if strings.Compare(text, "exit") == 0 || strings.Compare(text, "quit") == 0 {
			bot.Stop()
			return
		} else {
			mainLogger.Warn("error command!")
		}
	}
}

//程序结束时停止并释放bot
func waitExit(bot *Bot) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	mainLogger.Info("停止赛博监控")
	bot.Stop()
}

//定时器，在指定时间汇总数据
func summarize(bot *Bot, h, m int) {
	tick := time.Tick(time.Minute)
	for t := range tick {
		if (h == -1 || t.Hour() == h) && t.Minute() == m {
			fileName := bot.Summarize()
			if strings.Compare("", fileName) != 0 {
				bot.ReportSummarize(fileName)
			}
		}
	}
}

//推送错误消息，如果推送失败，写入到日志中
func pushAndLog(l *logger.Logger, msg string, args ...any) {
	go func() {
		err := pusher.Push(msg, args...)
		if err != nil {
			l.Error("推送消息失败，%v", err)
		}
		l.Debug("推送消息成功")
	}()
}

//读取设置信息，设置文件为 setting.json
func readSetting() (BotAccount, MonitorAccount, Board, config) {
	botAcc := BotAccount{}
	acc := MonitorAccount{}
	board := Board{}
	con := config{}
	settingFile, err := os.Open("setting.json")
	if err != nil {
		mainLogger.Error("读取设置失败，%v", err)
		panic(err)
	}
	defer settingFile.Close()
	data, err := io.ReadAll(settingFile)
	if err != nil {
		mainLogger.Error("%v", err)
		panic(err)
	}
	setting := gjson.ParseBytes(data)
	//登录账号所需要的cookie
	botAcc.uid = setting.Get("botAccount.uid").Uint()             //DedeUserID
	botAcc.uidMd5 = setting.Get("botAccount.uidMd5").String()     //DedeUserID__ckMd5
	botAcc.sessData = setting.Get("botAccount.sessData").String() //SESSDATA
	botAcc.csrf = setting.Get("botAccount.csrf").String()         //bili_jct
	botAcc.sid = setting.Get("botAccount.sid").String()           //sid

	//监控的账号
	acc.uid = setting.Get("account.uid").Uint()       //uid
	acc.alias = setting.Get("account.alias").String() //别名

	//评论区信息
	board.name = setting.Get("board.name").String() //别名
	//did, 例如：https://t.bilibili.com/662016827293958168 中的 662016827293958168 即是对应的did
	board.dId = setting.Get("board.oid").Uint()

	//每隔 freshCD 秒获取一次评论，值太小可能会被b站 ban ip
	con.freshCD = int(setting.Get("config.fresh").Int())
	con.likeCD = float32(setting.Get("config.like").Float()) //点赞一次后等待的秒数
	con.isLike = setting.Get("config.isLike").Bool()
	con.isPost = setting.Get("config.isPost").Bool()
	con.isFans = setting.Get("config.isFans").Bool()     //是否监控粉丝数变化
	con.hour = int(setting.Get("config.hour").Int())     //生成数据汇总的小时数，为 -1 则每小时生成一次
	con.minute = int(setting.Get("config.minute").Int()) //生成数据汇总的分钟数
	con.dbname = setting.Get("config.dbname").String()   //sqlite3 数据库名称，一个文件名即可

	loggerLevel := setting.Get("logger.level").String()       //日志级别
	loggerAppender := setting.Get("logger.appender").String() //日志写入文件还是直接在控制台输出

	//使用钉钉机器人推送消息，如果webhook为空字符串，则不会推送
	webhook := setting.Get("push.webhook").String()
	secret := setting.Get("push.secret").String()
	pusher = push.NewDingPusher(webhook, secret)

	switch loggerLevel {
	case "Debug":
		logLevel = logger.Debug
	case "Info":
		logLevel = logger.Info
	case "Warn":
		logLevel = logger.Warn
	case "Error":
		logLevel = logger.Error
	default:
		break
	}

	switch loggerAppender {
	case "file":
		logDst = logger.NewFileAppender(logFileSize)
	case "console":
		logDst = logger.NewConsoleAppender()
	default:
		break
	}
	return botAcc, acc, board, con
}
