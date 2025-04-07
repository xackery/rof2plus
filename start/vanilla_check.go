package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xackery/rof2plus/check"
	"github.com/xackery/rof2plus/checksum"
	"github.com/xackery/rof2plus/config"
)

// vanillaCheck checks if rof2 and ls are properly set, installed,
// and walks through process if not
func vanillaCheck() error {
	err := checkVanillaClient("rof2")
	if err != nil {
		return fmt.Errorf("checkRoF2: %w", err)
	}
	err = checkVanillaClient("ls")
	if err != nil {
		return fmt.Errorf("checkLS: %w", err)
	}

	return nil
}

func checkVanillaClient(client string) error {
	cfg := config.Get()
	path := ""
	switch client {
	case "rof2":
		path = cfg.RoF2Path
	case "ls":
		path = cfg.LSPath
	}

	if path != "" {
		fi, err := os.Stat(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("stat: %w", err)
			}
			path = ""
		}
		if !fi.IsDir() {
			path = ""
		}
	}

	if path != "" {
		return nil
	}

	fmt.Printf("I do not see %s installed. Do you have a vanilla copy of %s? (y/n) ", client, client)

	var answer string
	_, err := fmt.Scanln(&answer)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		return fmt.Errorf("invalid answer")
	}

	if strings.ToLower(answer) == "y" {
		isClientPathProvided := false
		for !isClientPathProvided {
			fmt.Printf("Please enter the path to your %s installation: ", client)
			_, err := fmt.Scanln(&path)
			if err != nil {
				fmt.Printf("scan path: %s", err)
				continue
			}
			path = strings.TrimSpace(path)
			if path == "" {
				fmt.Println("You must provide a valid path.")
				continue
			}

			fi, err := os.Stat(path)
			if err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("stat path: %w", err)
				}
				fmt.Println("Path does not exist. Please try again.")
				continue
			}

			if !fi.IsDir() {
				fmt.Println("Path is not a directory. Please try again.")
				continue
			}

			path = strings.TrimSuffix(path, string(os.PathSeparator))
			isClientPathProvided = true
			break
		}
	}

	isDepotDownloaded := false
	if strings.ToLower(answer) == "n" {
		fmt.Printf("You can install %s from Steam using the steam console.\n", client)
		fmt.Printf("Would you like me to open the console for you? (y/n) ")
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return fmt.Errorf("invalid answer")
		}
		if strings.ToLower(answer) == "y" {
			if runtime.GOOS == "windows" {
				cmd := exec.Command("cmd", "/c", "start", "steam://open/console")
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("exec steam (windows): %w", err)
				}
			} else if runtime.GOOS == "linux" {
				cmd := exec.Command("xdg-open", "steam://open/console")
				err = cmd.Run()
				if err != nil {
					return fmt.Errorf("exec steam (linux): %w", err)
				}
			} else {
				return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
			}

			depotCommand := "download_depot 205710 205711 1926608638440811669"
			if client == "ls" {
				depotCommand = "download_depot 205710 205711 5852850381673064693"
			}
			fmt.Printf("Please enter the following command to the console:\n%s\n", depotCommand)
			fmt.Println("I'll watch for steam to start downloading...")

			chkClient := checksum.ClientRoF2
			if client == "ls" {
				chkClient = checksum.ClientLS
			}
			path, err = monitorDepotDownload(chkClient)
			if err != nil {
				return fmt.Errorf("monitor depot: %w", err)
			}

			fmt.Println("Download complete!", client, "files are ready.")

		}
	}

	if isDepotDownloaded {
		fmt.Printf("It looks like your %s path is in steamapps.\n", client)
		fmt.Printf("I can move it to rof2plus\\%s. This is recommended.\n", client)
		fmt.Printf("Would you like me to do this? (y/n) ")
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			return fmt.Errorf("invalid answer")
		}
		if strings.ToLower(answer) == "y" {
			// Copy the depot files to the rof2plus directory
			err = os.MkdirAll(client, 0755)
			if err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}

			fmt.Printf("Moving depot files to %s directory...\n", client)
			err = os.Rename(path, client)
			if err != nil {
				return fmt.Errorf("rename: %w", err)
			}
			fmt.Println("Done.")
		}
		path = client
	}

	switch client {
	case "rof2":
		cfg.RoF2Path = path
	case "ls":
		cfg.LSPath = path
	}

	err = config.Get().Save()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	err = validateVanillaClient(client, path)
	if err != nil {
		switch client {
		case "rof2":
			cfg.RoF2Path = ""
		case "ls":
			cfg.LSPath = ""
		}
		config.Get().Save()
		return checkVanillaClient(client)

	}

	return nil
}

func validateVanillaClient(client string, path string) error {
	if client != "rof2" && client != "ls" {
		return fmt.Errorf("invalid client")
	}

	cl := checksum.ClientRoF2
	if client == "ls" {
		cl = checksum.ClientLS
	}
	checksum.SetClientLimit(true)

	err := check.Check(cl, path)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}

	report := check.Report()
	if report == nil {
		fmt.Println("Client is valid")
		return nil
	}

	if report.FailTotal > 0 {
		firstFail := report.Failures[0]
		if len(report.Failures) > 1 {

			fmt.Printf("Client %s is invalid, %d files failed.\n", client, report.FailTotal)
			fmt.Println("First failed file:", firstFail)
			fmt.Println("Please verify your installation")
			return fmt.Errorf("client is invalid")
		}
		fmt.Println("Note 1 file failed", firstFail, "but this is likely minor, ignoring")
	}

	return nil
}

// monitorDepotDownload checks if depotPath exists and monitors for xul.dll to reach >1MB
func monitorDepotDownload(client checksum.ChecksumClient) (string, error) {
	var err error
	depotPath := ""

	isWaiting := false
	// Check if directory exists, wait if it doesn't
	for {
		depotPath, err = SteamPath()
		if err != nil {
			if isWaiting {
				fmt.Printf(".")
				time.Sleep(6 * time.Second)
				continue
			}
			time.Sleep(12 * time.Second)
			continue
		}
		break
	}

	fmt.Println("Depot path:", depotPath)
	fmt.Println("Download started, I'll report key percents updates until finished")

	// Monitor for xul.dll with size > 1MB
	type watchFileEntry struct {
		name           string
		progressReport int
	}

	knownProgress := 0
	watchFiles := []watchFileEntry{
		{name: "10annvshield.eqg", progressReport: 1},
		{name: "B09.eqg", progressReport: 2},
		{name: "cabeast.s3d", progressReport: 5},
		{name: "cobaltscar_sounds.eff", progressReport: 6},
		{name: "crystalshard_obj.s3d", progressReport: 7},
		{name: "drachnidhivec.eqg", progressReport: 8},
		{name: "eastkorlach.eqg", progressReport: 9},
		{name: "eviltree.emt", progressReport: 10},
		{name: "frd.eqg", progressReport: 11},
		{name: "furniture09.eqg", progressReport: 12},
		{name: "global_chr.s3d", progressReport: 15},
		{name: "guildlobby.eqg", progressReport: 18},
		{name: "hole.s3d", progressReport: 20},
		{name: "iceclad.s3d", progressReport: 23},
		{name: "karnor.s3d", progressReport: 25},
		{name: "kerraridge.s3d", progressReport: 30},
		{name: "load_obj.s3d", progressReport: 35},
		{name: "maiden.s3d", progressReport: 38},
		{name: "mistythicket.eqg", progressReport: 40},
		{name: "northro.eqg", progressReport: 43},
		{name: "nro.s3d", progressReport: 45},
		{name: "oggok.s3d", progressReport: 48},
		{name: "paineel.s3d", progressReport: 50},
		{name: "potactics.s3d", progressReport: 55},
		{name: "powater.s3d", progressReport: 58},
		{name: "qeynos2.s3d", progressReport: 60},
		{name: "qeytoqrg.s3d", progressReport: 63},
		{name: "runnyeye.s3d", progressReport: 65},
		{name: "sirens.s3d", progressReport: 68},
		{name: "skyfire.s3d", progressReport: 70},
		{name: "soldungb.s3d", progressReport: 73},
		{name: "southkarana.s3d", progressReport: 75},
		{name: "swampofnohope.s3d", progressReport: 80},
		{name: "tenebrous.s3d", progressReport: 83},
		{name: "thurgadina.s3d", progressReport: 85},
		{name: "tutorialb.s3d", progressReport: 90},
		{name: "velketor.s3d", progressReport: 93},
		{name: "wallofslaughter.eqg", progressReport: 95},
		{name: "xorbb.eqg", progressReport: 98},
		{name: "zonein.eqg", progressReport: 100},
	}

	for {
		for _, watchFile := range watchFiles {
			if knownProgress >= watchFile.progressReport {
				continue
			}

			filePath := filepath.Join(depotPath, watchFile.name)
			fi, err := os.Stat(filePath)
			if err != nil {
				time.Sleep(2 * time.Second)
				break
			}

			if fi.Size() < checksum.FileSize(client, filepath.Base(filePath))-1000 {
				break
			}

			knownProgress = watchFile.progressReport
			fmt.Println("Download progress:", knownProgress, "%")
			if knownProgress == 100 {
				break
			}
		}
		if knownProgress == 100 {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return depotPath, nil
}
