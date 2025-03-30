package start

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xackery/rof2plus/check"
	"github.com/xackery/rof2plus/checksum"
	"github.com/xackery/rof2plus/serverlist"
)

func patch(server *serverlist.ServerEntry) error {
	fmt.Println("Downloading patch from:", server.PatchURL)

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

	err = check.Check(checksum.ClientRoF2, eqPath)
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}

	report := check.Report()
	if report != nil {
		if len(report.Failures) > 0 {
			fmt.Printf("Total: %d OK: %d Fail: %d (First Failure: Client: %s, Path: %s, Reason: %s)\n", report.FileTotal, report.OKTotal, report.FailTotal, report.Failures[0].Client.String(), report.Failures[0].Path, report.Failures[0].Directions)
			return fmt.Errorf("check failed")
		}
	} else {
		fmt.Println("No issues found.")
	}

	return nil
}
