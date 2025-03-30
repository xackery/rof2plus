package start

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// installCheck verifies you are not running the program in Downloads, Desktop, or other common generic folders
// if you are, it'll ask you to install it
func installCheck() error {
	// TODO: registry check for active installation

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get absolute path: %w", err)
	}

	notInstalledFolders := []string{
		"Desktop",
		"Documents",
		"Downloads",
	}
	for _, folder := range notInstalledFolders {
		if strings.Contains(exePath, folder) {
			fmt.Printf("It looks like you are running the program from your %s folder.\n", folder)
			fmt.Printf("Where would you like to install the program? ")
			path := ""
			_, err := fmt.Scanln(&path)
			if err != nil {
				return fmt.Errorf("scan path: %w", err)
			}
			path = strings.TrimSpace(path)
			if path == "" {
				fmt.Println("You must provide a valid path.")
				return fmt.Errorf("invalid path")
			}
			fi, err := os.Stat(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("stat path: %w", err)
				}
				err = os.MkdirAll(path, 0755)
				if err != nil {
					return fmt.Errorf("mkdir path: %w", err)
				}
			}
			if err == nil && !fi.IsDir() {
				return fmt.Errorf("path is not a directory")
			}
			outPath := filepath.Join(path, "rof2plus.exe")
			fi, err = os.Stat(outPath)
			if err == nil && !fi.IsDir() {
				fmt.Printf("File %s already exists. Overwrite? (y/n): ", outPath)
				var overwrite string
				_, err := fmt.Scanln(&overwrite)
				if err != nil {
					return fmt.Errorf("scan overwrite: %w", err)
				}
				if strings.ToLower(overwrite) != "y" {
					fmt.Println("Installation cancelled.")
					return nil
				}
			}
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("stat outPath: %w", err)
			}
			if err == nil {
				err = os.Remove(outPath)
				if err != nil {
					return fmt.Errorf("remove outPath: %w", err)
				}
			}

			w, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("create exe: %w", err)
			}
			defer w.Close()
			exeData, err := os.ReadFile(exePath)
			if err != nil {
				return fmt.Errorf("read exe: %w", err)
			}
			_, err = w.Write(exeData)
			if err != nil {
				return fmt.Errorf("write exe: %w", err)
			}
			fmt.Printf("Installed to %s\n", outPath)
			err = os.Chdir(path)
			if err != nil {
				return fmt.Errorf("chdir: %w", err)
			}
			fmt.Printf("Changed working directory to %s\n", path)
		}
	}

	return nil
}
