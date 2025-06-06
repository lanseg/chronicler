package main

import (
	"encoding/json"
	"os"

	cmd "chronicler/command"
	"chronicler/common"

	cfg "github.com/lanseg/goconfig"
)

var (
	logger   = common.NewLogger("main")
	commands = map[string]cmd.Command{
		"list":   cmd.List,
		"save":   cmd.Save,
		"view":   cmd.View,
		"export": cmd.Export,
	}
)

func getDefaultSettings(src string) (*cmd.Settings, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}
	result := &cmd.Settings{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func main() {
	settings, err := getDefaultSettings("settings.json")
	if err != nil {
		logger.Errorf("Cannot load default settings from file: %s. Zero values will be used", err)
		settings = &cmd.Settings{}
	}
	forFlag := &cfg.FlagSource{}
	cfg, err := cfg.GetConfigTo(settings, cfg.FromEnv, forFlag.Collect)
	if err != nil {
		logger.Errorf("Cannot load settings: %s", err)
		return
	}
	args := forFlag.Args()
	if len(args) == 0 {
		logger.Errorf("No command specified, available commands: %s", "COMMANDS")
		os.Exit(-1)
	}
	command, ok := commands[args[0]]
	if !ok {
		logger.Debugf("Unknown command (args %q)", os.Args)
		return
	}
	logger.Debugf("Running command %q with args %q", os.Args[1], os.Args[2:])
	if err := command(cfg, args[1:]); err != nil {
		logger.Errorf("Error: %s", err)
	}
}
