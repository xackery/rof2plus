package main

import (
	"fmt"
	"runtime"
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
	// Do something
	return nil
}
