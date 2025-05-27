package main

import (
	"os"

	cmd "chronicler/command"
	"chronicler/common"

	"github.com/lanseg/goconfig"
)

var (
	logger = common.NewLogger("main")
)

func getCommand(cmdName string) cmd.Command {
	switch cmdName {
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
	forFlag := &goconfig.FlagSource{}
	cfg, err := goconfig.GetConfig[cmd.Settings](goconfig.FromEnv, forFlag.Collect)
	if err != nil {
		logger.Errorf("Cannot load settings: %s", err)
		return
	}
	args := forFlag.Args()
	if len(args) == 0 {
		logger.Errorf("No command specified, available commands: %s", "COMMANDS")
		os.Exit(-1)
	}
	command := getCommand(args[0])
	if command == nil {
		logger.Debugf("Unknown command (args %q)", os.Args)
		return
	}
	logger.Debugf("Running command %q with args %q", os.Args[1], os.Args[2:])
	if err := command(cfg, args[1:]); err != nil {
		logger.Errorf("Error: %s", err)
	}
}
