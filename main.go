package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/bwiggs/go-nexrad/render"
)

func main() {

	// trace.Start(os.Stdout)
	// defer trace.Stop()

	logrus.SetLevel(logrus.InfoLevel)

	// f, _ := os.Open("data/KCRP20170826_000304_V06") // Harvey Approach
	// f, _ := os.Open("data/KCRP20170826_010340_V06") //Harvey Approach
	// f, _ := os.Open("data/KCRP20170826_020412_V06") // Harvey Approach
	// f, _ := os.Open("data/KCRP20170826_030439_V06") // Harvey Approach
	// f, _ := os.Open("data/KCRP20170826_040419_V06") // Harvey Approach
	f, _ := os.Open("data/KCRP20170826_050221_V06") // Harvey Landfall

	// f, _ := os.Open("data/KCRP20170826_235827_V06") // Harvey Central Texas

	defer f.Close()
	ar2 := archive2.Extract(f)
	render.DrawSequential(ar2.ElevationScans[1])
}
