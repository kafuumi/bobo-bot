package main

import "github.com/Hami-Lemon/bobo-bot/logger"

var (
	logLevel = logger.Debug
	//logDst     = logger.NewFileAppender(1024)
	logDst     = logger.NewConsoleAppender()
	mainLogger = logger.New("main", logLevel, logDst)
)
