package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"chronist"
	"chronist/storage"
	"chronist/telegram"
	"chronist/util"
)

var (
	telegramBotKey = flag.String(tgBotKeyFlag, "", "A key for the telegram bot api.")
	storageRoot    = flag.String(storageRootFlag, "chronist_storage", "A local folder to save downloads.")
)

const (
	privateChatId   = int64(0)
	tgBotKeyFlag    = "telegram_bot_key"
	storageRootFlag = "storage_root"
)

func getCursor() int64 {
	bytes, _ := os.ReadFile("cursor.txt")
	num, _ := strconv.Atoi(string(bytes))
	return int64(num)
}

func saveCursor(cursor int64) {
	os.WriteFile("cursor.txt", []byte(fmt.Sprintf("%d", cursor)), 0644)
}

func main() {
	flag.Parse()
	logger := util.NewLogger("main")

	if len(*telegramBotKey) == 0 {
		logger.Errorf("No telegram bot key defined, please set it with --%s=\"...\"", tgBotKeyFlag)
		return
	}

	chr := chronist.NewChronist(
		getCursor(),
		telegram.NewBot(*telegramBotKey),
		storage.NewStorage(*storageRoot),
	)

	newRequests, err := chr.FetchRequests()
	if err != nil {
		logger.Errorf("Cannot fetch the requests: %s", err.Error())
		return
	}
	chr.SaveRequests(newRequests)
	saveCursor(chr.GetCursor() + 1)
}
