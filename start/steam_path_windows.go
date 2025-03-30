// build windows
package start

import (
	"fmt"

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
	return steamPathValue, nil
}
