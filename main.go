package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/xackery/rof2plus/check"
	"github.com/xackery/rof2plus/checksum"
)

func main() {
	err := run()
	if err != nil {
		if runtime.GOOS == "windows" {
			fmt.Println("Press any key to exit...")
			fmt.Scanln()
		}
		fmt.Println("Failed to run the program: ", err)
	}
}

func run() error {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rof2plus [check] <pathToEQ>")
		os.Exit(1)
	}

	action := os.Args[1]
	path := os.Args[2]
	if strings.EqualFold(action, "check") {
		err := check.Check(checksum.ClientRoF2, path)
		if err != nil {
			return fmt.Errorf("check: %w", err)
		}

		report := check.Report()
		for _, failure := range report.Failures {
			fmt.Println(failure)
		}
		return nil
	}

	return nil
}
