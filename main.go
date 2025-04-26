package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	cmd "chronicler/command"
	"chronicler/common"
)

var (
	logger = common.NewLogger("main")
)

func getCommand() cmd.Commmand {
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

	flag.Parse()

	f, err := os.Create("cpuprofile")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	command := getCommand()
	if command == nil {
		logger.Debugf("Unknown command (args %q)", os.Args)
		return
	}
	logger.Debugf("Running command %q with args %q", os.Args[1], os.Args[2:])
	if err := command(cmd.GetSettings(), os.Args[2:]); err != nil {
		logger.Errorf("Error: %s", err)
	}

	ff, err := os.Create("memprofile")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer ff.Close() // error handling omitted for example
	runtime.GC()     // get up-to-date statistics
	// Lookup("allocs") creates a profile similar to go test -memprofile.
	// Alternatively, use Lookup("heap") for a profile
	// that has inuse_space as the default index.
	if err := pprof.Lookup("allocs").WriteTo(ff, 0); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
}
