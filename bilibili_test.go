package main

import (
	"github.com/tidwall/gjson"
	"io"
	"os"
	"testing"
)

var (
	bili *BiliBili
)

func init() {
	file, err := os.Open("./cookie.json")
	if err != nil {
		panic(err)
	}
	data, _ := io.ReadAll(file)
	result := gjson.ParseBytes(data)
	bot := BotAccount{
		Account: Account{
			uid: result.Get(DedeUserID).Uint(),
		},
		uidMd5:   result.Get(DedeUserIDMd5).String(),
		sessData: result.Get(SessData).String(),
		csrf:     result.Get(Csrf).String(),
		sid:      result.Get(SId).String(),
	}
	bili = BiliBiliLogin(bot)
}

func TestBiliBili_AccountInfo(t *testing.T) {
	println(bili.user.uname)
}
