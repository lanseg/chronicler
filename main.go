package main

import (
	"flag"
	"sync"

	"chronicler"
	"chronicler/storage"
	"chronicler/telegram"
	"chronicler/util"

	rpb "chronicler/proto/records"
)

func main() {
	flag.Parse()
	cfg := chronicler.GetConfig()
	stg := storage.NewStorage(*cfg.StorageRoot)
	ts := chronicler.NewTelegramChronicler(telegram.NewBot(*cfg.TelegramBotKey))

	go func() {
		log := util.NewLogger("Record loader")
		log.Infof("Starting record loader")
		for {
			recordSet := <-ts.GetRecords()
			if err := stg.SaveRecords(recordSet); err != nil {
				log.Warningf("Error while saving a record: %s", err)
				ts.SendResponse() <- &rpb.Response{Source: recordSet.Request.Source, Content: err.Error()}
			} else {
				ts.SendResponse() <- &rpb.Response{Source: recordSet.Request.Source, Content: "Saved"}
			}
		}
	}()

	go func() {
		log := util.NewLogger("Request loader")
		log.Infof("Starting request loader")
		for {
			request := <-ts.GetRequest()
			log.Infof("Got new request: %s", request)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}
