// build !windows
package start

import (
	"fmt"
	"os"
	"path/filepath"
)

// SteamPath returns the Steam path on Linux
func SteamPath() (string, error) {
	// Check common Steam installation paths
	paths := []string{
		"$HOME/.steam/steam",
		"$HOME/.steam/steam/ubuntu12_32",
		"$HOME/.local/share/Steam",
	}

	for _, path := range paths {
		expandedPath := os.ExpandEnv(path)

		depotPath := filepath.Join(expandedPath, "steamapps", "content", "app_205710", "depot_205711")
		if !checkPath(depotPath) {
			continue
		}

		return depotPath, nil
	}

	return "", fmt.Errorf("steam path not found on this system")
}

func checkPath(depotPath string) bool {
	fi, err := os.Stat(depotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	if !fi.IsDir() {
		return false
	}
	return true
}
