package command

import (
	"path/filepath"

	"chronicler/common"
	opb "chronicler/proto"
	"chronicler/storage"
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
