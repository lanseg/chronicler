package main

import (
	"encoding/json"
	"maps"
	"os"
	"slices"

	cmd "chronicler/command"
	"chronicler/common"

	cfg "github.com/lanseg/goconfig"
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

func mainf() int {
	logger := common.NewLogger("main")
	commands := map[string]cmd.Command{
		"list":   cmd.List,
		"save":   cmd.Save,
		"view":   cmd.View,
		"export": cmd.Export,
	}

	settings, err := getDefaultSettings("settings.json")
	if err != nil {
		logger.Infof("Cannot load default settings from file: %s. Will use only user input", err)
		settings = &cmd.Settings{}
	}
	forFlag := &cfg.FlagSource{}
	cfg, err := cfg.GetConfigTo(settings, cfg.FromEnv, forFlag.Collect)
	if err != nil {
		logger.Errorf("Cannot read configs: %s", err)
		return -1
	}

	args := forFlag.Args()
	if len(args) == 0 {
		logger.Infof("No command specified, supported commands are:")
		for _, cmd := range slices.Sorted(maps.Keys(commands)) {
			logger.Infof("\t%s", cmd)
		}
		return -2
	}

	command, ok := commands[args[0]]
	if !ok {
		logger.Errorf("Unknown command (args %q)", os.Args)
		return -3
	}

	logger.Debugf("Running command %q with args %q", args[0], args)
	if err := command(cfg, args[1:]); err != nil {
		logger.Errorf("Error: %s", err)
		return -4
	}
	return 0
}

func main() {
	os.Exit(mainf())
}
