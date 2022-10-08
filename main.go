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
  storageRoot = flag.String(storageRootFlag, "chronist_storage", "A local folder to save downloads.")
)

const (
	privateChatId = int64(0)
    tgBotKeyFlag = "telegram_bot_key"
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

    if len(*telegramBotKey) == 0{
      logger.Errorf("No telegram bot key defined, please set it with --%s=\"...\"", tgBotKeyFlag)
      return
    }

    tg := telegram.NewBot(*telegramBotKey)
	stg := storage.NewStorage(*storageRoot)
	chr := chronist.NewChronist(getCursor(), tg)

	newRequests, err := chr.FetchRequests()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	requestBySource := util.GroupBy(newRequests, func(r *storage.Record) *storage.Source {
		return r.Source
	})
	for src, reqs := range requestBySource {
		success := []*storage.Record{}
		failure := []*storage.Record{}
		for _, req := range reqs {
			if err := stg.SaveRecord(req); err != nil {
				failure = append(failure, req)
				logger.Errorf("failed to save record %v: %s\n", req, err)
			} else {
				success = append(success, req)
				logger.Infof("Saved record %v\n", req)
			}
		}
		id, _ := strconv.Atoi(src.ChannelID)
		tg.SendMessage(int64(id),
			fmt.Sprintf("Saved %d new records, failed to save: %d",
				len(success), len(failure)))
	}
	saveCursor(chr.GetCursor() + 1)
}
