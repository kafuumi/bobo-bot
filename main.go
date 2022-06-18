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
	//logDst   = logger.NewFileAppender(1024 * 512)
	logDst     = logger.NewConsoleAppender()
	mainLogger = logger.New("main", logLevel, logger.NewConsoleAppender())
)

func main() {
	botAccount, account, board := readSetting()
	bili := BiliBiliLogin(botAccount)
	if bili == nil {
		mainLogger.Error("登录失败！")
		return
	} else {
		mainLogger.Info("登录成功，%s", bili.user.uname)
	}
	ma := MonitorAccount{
		Account: account,
	}
	board.Account = account
	bot := NewBot(bili, board, ma, 10, 1)
	go waitExit(bot)
	go summarize(bot, 7, 33)
	mainLogger.Info("开始赛博监控...")
	mainLogger.Info("监控评论区：%s, %d", board.name, board.oid)
	defer logDst.Close()
	bot.Monitor()
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

func readSetting() (BotAccount, Account, Board) {
	botAcc := BotAccount{}
	acc := Account{}
	board := Board{}
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
	botAcc.uid = setting.Get("botAccount.uid").Uint()
	botAcc.uidMd5 = setting.Get("botAccount.uidMd5").String()
	botAcc.sessData = setting.Get("botAccount.sessData").String()
	botAcc.csrf = setting.Get("botAccount.csrf").String()
	botAcc.sid = setting.Get("botAccount.sid").String()

	acc.uid = setting.Get("account.uid").Uint()
	acc.alias = setting.Get("account.alias").String()

	board.name = setting.Get("board.name").String()
	board.oid = setting.Get("board.oid").Uint()
	return botAcc, acc, board
}
