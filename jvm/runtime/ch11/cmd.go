package main

import "flag"
import "fmt"
import "os"

// java [-options] class [args...]
type Cmd struct {
	helpFlag         bool
	versionFlag      bool
	verboseClassFlag bool // 是否把类加载信息输出到控制台
	verboseInstFlag  bool // 是否把指令执行信息输出到控制台
	sandboxFlag      bool
	maxInstructions  uint64
	maxHeapBytes     uint64
	maxArrayLength   uint64
	maxOutputBytes   uint64
	cpOption         string
	XjreOption       string
	class            string
	args             []string
}

func parseCmd() *Cmd {
	cmd := &Cmd{}

	flag.Usage = printUsage
	flag.BoolVar(&cmd.helpFlag, "help", false, "print help message")
	flag.BoolVar(&cmd.helpFlag, "?", false, "print help message")
	flag.BoolVar(&cmd.versionFlag, "version", false, "print version and exit")
	flag.BoolVar(&cmd.verboseClassFlag, "verbose", false, "enable verbose output")
	flag.BoolVar(&cmd.verboseClassFlag, "verbose:class", false, "enable verbose output")
	flag.BoolVar(&cmd.verboseInstFlag, "verbose:inst", false, "enable verbose output")
	flag.BoolVar(&cmd.sandboxFlag, "sandbox", false, "enable execution limits")
	flag.Uint64Var(&cmd.maxInstructions, "max-instructions", 0, "maximum bytecode instructions")
	flag.Uint64Var(&cmd.maxHeapBytes, "max-heap-bytes", 0, "maximum managed allocation bytes")
	flag.Uint64Var(&cmd.maxArrayLength, "max-array-length", 0, "maximum elements per array")
	flag.Uint64Var(&cmd.maxOutputBytes, "max-output-bytes", 0, "maximum stdout bytes")
	flag.StringVar(&cmd.cpOption, "classpath", "", "classpath")
	flag.StringVar(&cmd.cpOption, "cp", "", "classpath")
	flag.StringVar(&cmd.XjreOption, "Xjre", "", "path to jre")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		cmd.class = args[0]
		cmd.args = args[1:]
	}

	return cmd
}

func printUsage() {
	fmt.Printf("Usage: %s [-options] class [args...]\n", os.Args[0])
	//flag.PrintDefaults()
}
