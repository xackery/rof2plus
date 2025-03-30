package check

import (
	"fmt"
	"os"
	"testing"

	"github.com/xackery/rof2plus/checksum"
	"gopkg.in/yaml.v3"
)

func TestCheck(t *testing.T) {
	targetPath := os.Getenv("TARGET_PATH")
	if targetPath == "" {
		t.Skip("TARGET_PATH not set")
	}
	err := Check(checksum.ClientLS, targetPath)
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
	targetPath := os.Getenv("TARGET_PATH")
	if targetPath == "" {
		t.Skip("TARGET_PATH not set")
	}
	data, err := os.ReadFile("../bin/filelist_rof.yml")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	fileList := &checksum.FileList{}

	err = yaml.Unmarshal(data, fileList)
	if err != nil {
		t.Fatalf("failed to unmarshal filelist: %v", err)
	}

	err = checksum.SetPatcherFilelist(fileList)
	if err != nil {
		t.Fatalf("failed to set patcher filelist: %v", err)
	}

	checksum.SetExcludedClients(checksum.ClientLS)

	err = Check(checksum.ClientPatcher, targetPath)
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
