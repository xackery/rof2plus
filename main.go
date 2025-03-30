package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/xackery/rof2plus/check"
	"github.com/xackery/rof2plus/checksum"
	"github.com/xackery/rof2plus/config"
	"github.com/xackery/rof2plus/start"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run the program:", err)
		if runtime.GOOS == "windows" {
			fmt.Println("Press any key to exit...")
			fmt.Scanln()
		}

	}
}

func run() error {
	var err error
	if len(os.Args) < 2 {
		fmt.Println("Usage: rof2plus <start|check>")
		os.Exit(1)
	}

	action := os.Args[1]
	arg1 := ""
	if len(os.Args) >= 3 {
		arg1 = os.Args[2]
	}
	switch strings.ToLower(action) {
	case "start":
		err = start.Start(arg1)
		if err != nil {
			return fmt.Errorf("start: %w", err)
		}
	case "check":
		if arg1 == "" {
			fmt.Println("Usage: rof2plus check <path>")
			os.Exit(1)
		}
		_, err := config.New(context.Background(), "rof2plus")
		if err != nil {
			return fmt.Errorf("config.New: %w", err)
		}

		err = check.Check(checksum.ClientRoF2, arg1)
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
