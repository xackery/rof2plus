package start

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xackery/rof2plus/checksum"
	"github.com/xackery/rof2plus/config"
)

// vanillaCheck checks if rof2 and ls are properly set, installed,
// and walks through process if not
func vanillaCheck() error {
	cfg := config.Get()
	err := checkVanillaClient("rof2", &cfg.RoF2Path)
	if err != nil {
		return fmt.Errorf("checkRoF2: %w", err)
	}
	err = checkVanillaClient("ls", &cfg.LSPath)
	if err != nil {
		return fmt.Errorf("checkLS: %w", err)
	}

	fmt.Println("Saving changes")
	err = cfg.Save()
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

func checkVanillaClient(client string, pathStr *string) error {
	path := *pathStr

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
		pathStr = &path
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
			cmd := exec.Command("cmd", "/c", "start", "steam://open/console")
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("exec steam: %w", err)
			}

			depotCommand := "download_depot 205710 205711 1926608638440811669"
			if client == "ls" {
				depotCommand = "download_depot 205710 205711 5852850381673064693"
			}
			fmt.Printf("Please enter the following command to the console: %s.\n", depotCommand)
			fmt.Printf("Press Y and Enter once done.\n")
			_, err := fmt.Scanln(&answer)
			if err != nil {
				return fmt.Errorf("scan: %w", err)
			}

			// check the registry HKEY_CURRENT_USER\Software\Valve\Steam for key SteamPath

			steamPath, err := SteamPath()
			if err != nil {
				return fmt.Errorf("steam path: %w", err)
			}

			// Path to the downloaded depot will likely be in the Steam download folder
			depotPath := filepath.Join(steamPath, "steamapps", "content", "app_205710", "depot_205711")

			fmt.Println("I'm watching for the depot to be downloaded to your Steam library.")
			fmt.Println("This will monitor directory:", depotPath)
			fmt.Println("Press Ctrl+C to cancel at any time...")

			chkClient := checksum.ClientRoF2
			if client == "ls" {
				chkClient = checksum.ClientLS
			}
			err = monitorDepotDownload(chkClient, depotPath)
			if err != nil {
				return fmt.Errorf("monitor depot: %w", err)
			}

			path = depotPath
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

	pathStr = &path

	return nil
}

// monitorDepotDownload checks if depotPath exists and monitors for xul.dll to reach >1MB
func monitorDepotDownload(client checksum.ChecksumClient, depotPath string) error {
	// Check if directory exists, wait if it doesn't
	for {
		fi, err := os.Stat(depotPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("Waiting for download to begin...", depotPath)
				time.Sleep(5 * time.Second)
				continue
			}
			return fmt.Errorf("stat depot path: %w", err)
		}

		if !fi.IsDir() {
			return fmt.Errorf("depot path is not a directory")
		}

		break
	}

	fmt.Println("Download started, monitoring progress...")

	// Monitor for xul.dll with size > 1MB
	type watchFileEntry struct {
		name           string
		progressReport int
	}

	knownProgress := 0
	watchFiles := []watchFileEntry{
		{name: "!CheckMinSpec.dll", progressReport: 1},
		{name: "B09.eqg", progressReport: 2},
		{name: "HighPassHold.mp3", progressReport: 3},
		{name: "Tutorial.mp3", progressReport: 4},
		{name: "cabeast_chr.s3d", progressReport: 5},
		{name: "cobaltscar_sounds.eff", progressReport: 6},
		{name: "crystalshard_obj.s3d", progressReport: 7},
		{name: "drachnidhivec.eqg", progressReport: 8},
		{name: "eastkorlach.eqg", progressReport: 9},
		{name: "eviltree.emt", progressReport: 10},
		{name: "frd.eqg", progressReport: 11},
		{name: "furniture09.eqg", progressReport: 12},
		{name: "global_chr.s3d", progressReport: 15},
		{name: "guildlobby.eqg", progressReport: 18},
		{name: "hole.eqg", progressReport: 20},
		{name: "iceclad.eqg", progressReport: 23},
		{name: "karnor.eqg", progressReport: 25},
		{name: "kerraridge.eqg", progressReport: 28},
		{name: "lake_chr.s3d", progressReport: 30},
		{name: "lavastorm.eqg", progressReport: 33},
		{name: "load_example01.tga", progressReport: 35},
		{name: "maiden.eqg", progressReport: 38},
		{name: "misty.eqg", progressReport: 40},
		{name: "northro.eqg", progressReport: 43},
		{name: "nro.eqg", progressReport: 45},
		{name: "oggok.eqg", progressReport: 48},
		{name: "paineel.eqg", progressReport: 50},
		{name: "particle.eff", progressReport: 53},
		{name: "potactics.eqg", progressReport: 55},
		{name: "powater.eqg", progressReport: 58},
		{name: "qeynos2.eqg", progressReport: 60},
		{name: "qeytoqrg.eqg", progressReport: 63},
		{name: "runnyeye.eqg", progressReport: 65},
		{name: "sirens.eqg", progressReport: 68},
		{name: "skyfire.eqg", progressReport: 70},
		{name: "soldungb.eqg", progressReport: 73},
		{name: "southkarana.eqg", progressReport: 75},
		{name: "steamfactory.eqg", progressReport: 78},
		{name: "swampofnohope_chr.s3d", progressReport: 80},
		{name: "tenebrous_chr.s3d", progressReport: 83},
		{name: "thurgadina.eqg", progressReport: 85},
		{name: "tox.eqg", progressReport: 88},
		{name: "tutorialb.eqg", progressReport: 90},
		{name: "velketor.eqg", progressReport: 93},
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
			_, err := os.Stat(filePath)
			if err != nil {
				time.Sleep(2 * time.Second)
				break
			}

			md5, err := checksum.MD5Generate(filePath)
			if err != nil {
				time.Sleep(2 * time.Second)
				break
			}

			if md5 != checksum.MD5Hash(client, filepath.Base(filePath)) {
				break
			}

			knownProgress = watchFile.progressReport
			fmt.Println("Download progress:", knownProgress, "%")
			if knownProgress == 100 {
				break
			}
		}
		if knownProgress == 100 {
			fmt.Println("Download complete!")
			break
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}
