package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/bwiggs/go-nexrad/render"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	f, _ := os.Open("data/KCRP20170826_235827_V06")
	// f, _ := os.Open("data/KGRK20170827_234611_V06")
	defer f.Close()
	ar2 := archive2.Extract(f)
	render.Draw(ar2.ElevationScans[1])
}
