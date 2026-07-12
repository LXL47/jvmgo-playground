package main

import (
	"fmt"
	"os"

	"jvmgo/ch11/sandbox"
)

func main() {
	os.Exit(run())
}

func run() (exitCode int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Fprintln(os.Stderr, recovered)
			exitCode = 1
		}
	}()

	cmd := parseCmd()
	sandbox.Configure(sandbox.Limits{
		Enabled:         cmd.sandboxFlag,
		MaxInstructions: cmd.maxInstructions,
		MaxHeapBytes:    cmd.maxHeapBytes,
		MaxArrayLength:  cmd.maxArrayLength,
		MaxOutputBytes:  cmd.maxOutputBytes,
	})

	if cmd.versionFlag {
		println("version 0.0.1")
	} else if cmd.helpFlag || cmd.class == "" {
		printUsage()
	} else {
		newJVM(cmd).start()
	}
	return 0
}
