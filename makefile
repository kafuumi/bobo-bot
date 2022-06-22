ifeq ($(LANG),)
# for windows
build_time = $(shell echo %date:~0,4%-%date:~5,2%-%date:~8,2%,%time:~0,5%)
all:build
build:
	go build -ldflags="-s -w -X main.buildTime=$(build_time)" .
clean:
	del bobo-bot
	del bobo-bot.exe
else
# for linux
build_time =$(shell date -d now "+%Y-%m-%d,%H:%M")
all:build
build:
	go build -ldflags="-s -w -X main.buildTime=$(build_time)" .

upx:
	upx -9 bobo-bot
	upx -9 bobo-bot.exe
clean:
	rm ./bobo-bot
	rm ./bobo-bot.exe
endif