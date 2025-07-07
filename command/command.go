package command

import (
	"chronicler/common"
	"chronicler/exporter"
	opb "chronicler/proto"
	"chronicler/storage"
	"fmt"
	"path/filepath"
)

type Command = func(*Settings, []string) error

func getStorage(s *Settings, link string) (storage.Storage, error) {
	itemLink, err := opb.ParseLink(link)
	if err != nil {
		return nil, err
	}
	storage, err := storage.NewLocalStorage(filepath.Join(s.Storage.Root, common.UUID4For(itemLink)))
	if err != nil {
		return nil, err
	}
	return storage, nil
}

func Export(s *Settings, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Export requires two arguments: <link> and <target>, but got %q", args)
	}
	storage, err := getStorage(s, args[0])
	if err != nil {
		return err
	}
	return exporter.NewLocalExporter(storage).Export(args[1])
}

func View(s *Settings, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("View requires one argument: <link>, but got none")
	}
	storage, err := getStorage(s, args[0])
	if err != nil {
		return err
	}
	return exporter.NewTextExporter(storage).Export("")
}
