package check

import (
	"testing"

	"github.com/xackery/rof2plus/checksum"
)

func TestCheck(t *testing.T) {

	err := Check(checksum.ClientRoF2, "c:/games/eq/thj/")
	if err != nil {
		t.Fatalf("check: %v", err)
	}

}
