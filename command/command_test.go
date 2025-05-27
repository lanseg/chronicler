package command

import (
	"testing"
)

func makeSettings(t *testing.T) *Settings {
	return &Settings{
		Storage: &StorageSettings{
			Root: t.TempDir(),
		},
	}
}

func TestExportCommand(t *testing.T) {
	Export(makeSettings(t), []string{})
}

func TestListCommand(t *testing.T) {
	List(makeSettings(t), []string{})
}

func TestSaveCommand(t *testing.T) {
	Save(makeSettings(t), []string{})
}

func TestViewCommand(t *testing.T) {
	View(makeSettings(t), []string{})
}
