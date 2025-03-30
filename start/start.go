// start is responsible with starting the flow of rof2plus from cli
package start

import (
	"context"
	"fmt"

	"github.com/xackery/rof2plus/config"
	"github.com/xackery/rof2plus/serverlist"
)

// Start begins the program process
func Start(serverName string) error {
	err := installCheck()
	if err != nil {
		return fmt.Errorf("installCheck: %w", err)
	}

	_, err = config.New(context.Background(), "rof2plus")
	if err != nil {
		return fmt.Errorf("config.New: %w", err)
	}

	err = vanillaCheck()
	if err != nil {
		return fmt.Errorf("vanillaCheck: %w", err)
	}

	err = serverlist.Fetch()
	if err != nil {
		return fmt.Errorf("serverlist.fetch: %w", err)
	}

	server, err := selectServer(serverName)
	if err != nil {
		return fmt.Errorf("selectServer: %w", err)
	}

	if server == nil {
		return fmt.Errorf("no server selected")
	}

	fmt.Printf("Selected server: %s\n", server.Name)
	err = patchCheck(server)
	if err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	err = launch(server)
	if err != nil {
		return fmt.Errorf("launch: %w", err)
	}

	return nil
}
