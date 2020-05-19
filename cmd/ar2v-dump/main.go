package main

import (
	"fmt"
	"os"

	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/sirupsen/logrus"
)

func main() {
	f, err := os.Open("test.ar2v")
	logrus.SetLevel(logrus.DebugLevel)
	defer f.Close()
	if err != nil {
		logrus.Error(err)
		return
	}

	ar2 := archive2.Extract(f)

	fmt.Printf("Station: %s\n", ar2.VolumeHeader.ICAO)
	fmt.Printf("Date: %s\n", ar2.VolumeHeader.Date())
	fmt.Printf("File: %s\n", ar2.VolumeHeader.FileName())

	// spew.Dump(ar2.VolumeHeader)
}
