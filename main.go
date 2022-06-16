package main

import (
	"github.com/Hami-Lemon/bobo-bot/logger"
	"github.com/tidwall/gjson"
	"io"
	"os"
	"os/signal"
)

var (
	logLevel = logger.Info
	//logDst     = logger.NewFileAppender(1024)
	logDst     = logger.NewConsoleAppender()
	mainLogger = logger.New("main", logLevel, logDst)
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
	account := Account{
		uname: "",
		uid:   33605910,
		alias: "",
	}
	ma := MonitorAccount{
		Account:  account,
		follower: 0,
		face:     "",
		sign:     "",
	}
	board := Board{
		Account:  account,
		name:     "",
		oid:      671651306569465856,
		typeCode: 0,
		count:    0,
	}
	bot := NewBot(bili, board, ma, 5, 1)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)
		<-ch
		bot.Stop()
		mainLogger.Info("程序结束")
	}()
	mainLogger.Info("开始赛博监控...")
	bot.Monitor()
}
