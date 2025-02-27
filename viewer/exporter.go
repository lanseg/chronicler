package viewer

import (
	"chronicler/common"
	"chronicler/iferr"
	opb "chronicler/proto"
	"chronicler/storage"
	"path/filepath"
)

type Exporter struct {
	Root   string
	Target string
	logger *common.Logger
}

func NewExporter(root string, target string) *Exporter {
	return &Exporter{
		Root:   root,
		Target: target,
		logger: common.NewLogger("export"),
	}
}

func (v *Exporter) Export(id string) error {
	store := storage.BlockStorage{
		Storage: iferr.Exit(storage.NewLocalStorage(filepath.Join(v.Root, id))),
	}
	v.logger.Infof("Loading objects from %q", objectFileName)
	result := &opb.Snapshot{}
	if err := store.GetObject(&storage.GetRequest{Url: objectFileName}, &result); err != nil {
		return err
	}
	total := len(result.Objects)
	v.logger.Infof("Loaded objects: %d", total)
	for i, obj := range result.Objects {
		v.logger.Infof("[%06d of %06d] exporting %q to %q", i, total, obj.Id, v.Target)
	}
	return nil
}
