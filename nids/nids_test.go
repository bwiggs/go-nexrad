package nids

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestOpen(t *testing.T) {
	nids, err := Open("testdata/NIDS_DYX_NST_NST_20200522_0222")
	if err != nil {
		fmt.Printf("failed to open NIDS file: %s\n", err)
		t.Fail()
	}

	spew.Dump(nids)

}
