package main

import (
	"os"

	"chronicler/adapter/reddit"
	"chronicler/adapter/twitter"
	cmd "chronicler/command"
	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
)

var (
	logger = common.NewLogger("main")
)

type Settings struct {
	Twitter *twitter.Settings `json:"twitter"`
	Reddit  *reddit.Settings  `json:"reddit"`
	Storage *storage.Settings `json:"storage"`
}

func getSettings() *Settings {
	return &Settings{
		Twitter: &twitter.Settings{
			Token: os.Getenv("TWITTER_TOKEN"),
		},
		Reddit: &reddit.Settings{
			Token: os.Getenv("REDDIT_TOKEN"),
		},
		Storage: &storage.Settings{
			Root: os.Getenv("STORAGE_ROOT"),
		},
	}
}

func getCommand() func(*Settings, []string) {
	switch os.Args[1] {
	case "list":
		return list
	case "save":
		return save
	case "view":
		return view
	case "export":
		return export
	}
	return nil
}

func list(s *Settings, _ []string) {
	cmd.List(s.Storage.Root)
}

func view(s *Settings, args []string) {
	cmd.NewViewer(s.Storage.Root).View(common.UUID4For(common.OrExit(opb.ParseLink(args[0]))))
}

func export(s *Settings, args []string) {
	cmd.NewExporter(s.Storage.Root, args[0]).Export(common.UUID4For(common.OrExit(opb.ParseLink(args[0]))))
}

func save(s *Settings, args []string) {
	cmd.Save(s.Twitter, s.Reddit, s.Storage, args)
}

func main() {
	command := getCommand()
	if command == nil {
		logger.Debugf("Unknown command (args %q)", os.Args)
		return
	}
	logger.Debugf("Running command %q with args %q", os.Args[1], os.Args[2:])
	command(getSettings(), os.Args[2:])
}
