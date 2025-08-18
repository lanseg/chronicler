package command

import (
	"fmt"

	"chronicler/exporter"
)

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
