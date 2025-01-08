package check

import (
	"fmt"
	"os"
	"testing"

	"github.com/xackery/rof2plus/checksum"
)

func TestCheck(t *testing.T) {

	err := Check(checksum.ClientLS, "c:/games/eq/thj/")
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	report := Report()
	if report == nil {
		t.Fatalf("report is nil")
	}
	fmt.Printf("report: %+v\n", report)
	if report.FileTotal == 0 {
		t.Fatalf("report.FileTotal is 0")
	}
	if report.FailTotal == 0 {
		t.Fatalf("report.FailTotal is 0")
	}
	if report.OKTotal == 0 {
		t.Fatalf("report.OKTotal is 0")
	}

	t.Fatalf("report: %+v", report)

}

func TestPatcherCheck(t *testing.T) {
	data, err := os.ReadFile("../bin/filelist_rof.yml")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	err = checksum.SetPatcherFilelist(data)
	if err != nil {
		t.Fatalf("failed to set patcher filelist: %v", err)
	}

	checksum.SetExcludedClients(checksum.ClientLS, checksum.ClientLSOptional)

	err = Check(checksum.ClientPatcher, "c:/games/eq/thj-clean/")
	if err != nil {
		t.Fatalf("check: %v", err)
	}

	report := Report()
	if report == nil {
		t.Fatalf("report is nil")
	}
	fmt.Printf("report: %+v\n", report)
	if report.FileTotal == 0 {
		t.Fatalf("report.FileTotal is 0")
	}
	if report.FailTotal == 0 {
		t.Fatalf("report.FailTotal is 0")
	}
	if report.OKTotal == 0 {
		t.Fatalf("report.OKTotal is 0")
	}

	t.Fatalf("report: %+v", report)
}
