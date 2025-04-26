package main

import (
	"os"

	cmd "chronicler/command"
	"chronicler/common"
)

var (
	logger = common.NewLogger("main")
)

func getCommand() cmd.Command {
	switch os.Args[1] {
	case "list":
		return cmd.List
	case "save":
		return cmd.Save
	case "view":
		return cmd.View
	case "export":
		return cmd.Export
	}
	return nil
}

func main() {
	command := getCommand()
	if command == nil {
		logger.Debugf("Unknown command (args %q)", os.Args)
		return
	}
	logger.Debugf("Running command %q with args %q", os.Args[1], os.Args[2:])
	if err := command(cmd.GetSettings(), os.Args[2:]); err != nil {
		logger.Errorf("Error: %s", err)
	}
}
