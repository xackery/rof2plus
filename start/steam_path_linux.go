// build !windows
package start

import "fmt"

// SteamPath returns the steam path
func SteamPath() (string, error) {
	return "", fmt.Errorf("steam is not supported on non-windows for now")
}
