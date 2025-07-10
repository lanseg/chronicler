package exporter

import (
	"path/filepath"
	"testing"

	"chronicler/storage"
)

func TestHtmlExporter(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "snapshot without files", path: "test_data/snapshot_no_files"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewLocalStorage(tc.path)
			if err != nil {
				t.Errorf("Cannot create storage: %s", err)
				return
			}
			ex := HtmlExporter(s)
			if err := ex.Export(filepath.Join(t.TempDir(), "target")); err != nil {
				t.Errorf("Error while exporting data: %s", err)
			}
		})
	}
}

func TestTextExporter(t *testing.T) {

	for _, tc := range []struct {
		name string
	}{} {
		t.Run(tc.name, func(t *testing.T) {

		})
	}
}
