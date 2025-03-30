package start

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xackery/rof2plus/checksum"
	"github.com/xackery/rof2plus/patch"
	"github.com/xackery/rof2plus/serverlist"
)

func patchCheck(server *serverlist.ServerEntry) error {

	eqPath := filepath.Join(server.ShortName)

	fi, err := os.Stat(eqPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(eqPath, 0755)
			if err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
		} else {
			return fmt.Errorf("stat: %w", err)
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", eqPath)
	}

	fileList, err := checksum.FetchPatcherFilelist(server.PatchURL)
	if err != nil {
		return fmt.Errorf("fetch patcher filelist: %w", err)
	}

	err = patch.Download(fileList, eqPath)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	return nil
}
