package main

import (
	"github.com/Hami-Lemon/bobo-bot/logger"
	"github.com/tidwall/gjson"
	"io"
	"os"
	"os/signal"
	"time"
)

var (
	logLevel = logger.Info
	logDst   = logger.NewFileAppender(1024 * 512)
	//logDst     = logger.NewConsoleAppender()
	mainLogger = logger.New("main", logLevel, logger.NewConsoleAppender())
)

func main() {
	file, err := os.Open("./cookie.json")
	if err != nil {
		panic(err)
	}
	data, _ := io.ReadAll(file)
	result := gjson.ParseBytes(data)
	botAccount := BotAccount{
		Account: Account{
			uid: result.Get(DedeUserID).Uint(),
		},
		uidMd5:   result.Get(DedeUserIDMd5).String(),
		sessData: result.Get(SessData).String(),
		csrf:     result.Get(Csrf).String(),
		sid:      result.Get(SId).String(),
	}
	bili := BiliBiliLogin(botAccount)
	if bili == nil {
		mainLogger.Error("登录失败！")
		return
	} else {
		mainLogger.Info("登录成功，%s", bili.user.uname)
	}
	account := Account{
		uname: "",
		uid:   33605910,
		alias: "美女宝",
	}
	ma := MonitorAccount{
		Account:  account,
		follower: 0,
		face:     "",
		sign:     "",
	}
	board := Board{
		Account:  account,
		name:     "代版",
		oid:      672427609096716297,
		typeCode: 0,
		count:    0,
	}
	bot := NewBot(bili, board, ma, 10, 1)
	go waitExit(bot)
	go summarize(bot, 7, 33)
	mainLogger.Info("开始赛博监控...")
	defer logDst.Close()
	bot.Monitor()
	bot.Summarize()
}

func waitExit(bot *Bot) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
	bot.Stop()
	mainLogger.Info("停止赛博监控")
}

func summarize(bot *Bot, h, m int) {
	tick := time.Tick(time.Minute)
	for t := range tick {
		if (h == -1 || t.Hour() == h) && t.Minute() == m {
			bot.Summarize()
		}
	}
}
