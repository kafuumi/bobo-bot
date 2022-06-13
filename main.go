package main

import (
	"github.com/Hami-Lemon/bobo-bot/logger"
	"os"
	"os/signal"
)

var (
	logLevel = logger.Debug
	//logDst     = logger.NewFileAppender(1024)
	logDst     = logger.NewConsoleAppender()
	mainLogger = logger.New("main", logLevel, logDst)
)

func main() {
	WaitExit()
	println("exit")
}

func WaitExit() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
}
