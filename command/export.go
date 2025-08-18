package command

import (
	"fmt"

	"chronicler/exporter"
)

func Export(s *Settings, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Export requires two arguments: <link> and <target>, but got %q", args)
	}
	storage, err := getStorage(s, args[0])
	if err != nil {
		return err
	}
	return exporter.HtmlExporter(storage).Export(args[1])
}
