// build windows
package start

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// SteamPath returns the steam path
func SteamPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("steam registry: %w", err)
	}
	defer k.Close()

	steamPathValue, _, err := k.GetStringValue("SteamPath")
	if err != nil {
		return "", fmt.Errorf("steam path: %w", err)
	}

	depotPath := filepath.Join(steamPathValue, "steamapps", "content", "app_205710", "depot_205711")
	fi, err := os.Stat(depotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("depot path does not exist: %w", err)
		}
		return "", fmt.Errorf("stat depot path: %w", err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("depot path is not a directory")
	}

	return depotPath, nil
}
