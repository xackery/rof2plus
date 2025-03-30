package start

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xackery/rof2plus/serverlist"
)

func launch(server *serverlist.ServerEntry) error {

	err := os.Chdir(server.ShortName)
	if err != nil {
		return fmt.Errorf("chdir: %w", err)
	}

	cmd := exec.Command("eqgame.exe", "patchme")
	cmd.Dir = server.ShortName

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("start: %w", err)
	}

	return nil
}
