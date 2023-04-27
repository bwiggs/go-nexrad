package archive2

import (
	"os"
	"testing"
)

func TestExtract(t *testing.T) {
	//tamu, err := os.Open("testdata/TAMU_20200808_2058")
	tamu, err := os.Open("testdata/TAMU_20200807_2104")
	if err != nil {
		t.Fatal(err)
	}
	Extract(tamu)
}

func TestRawDatasets(t *testing.T) {
	files := []string{
		"KCRP20210919_000249_V06",
		"KGRK20200914_043239_V06",
	}

	for _, f := range files {
		t.Log("testing: " + f)
		tamu, err := os.Open("testdata/" + f)
		if err != nil {
			t.Fatal(err)
		}
		Extract(tamu)
	}
}
