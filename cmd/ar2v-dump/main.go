package main

import (
	"fmt"
	"os"

	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/sirupsen/logrus"
)

func main() {
	f, err := os.Open(os.Args[1])
	logrus.SetLevel(logrus.InfoLevel)

	defer f.Close()
	if err != nil {
		logrus.Error(err)
		return
	}

	ar2, err := archive2.Extract(f)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("Station: %s\n", ar2.VolumeHeader.ICAO)
	fmt.Printf("Date: %s\n", ar2.VolumeHeader.Date())
	fmt.Printf("File: %s\n", ar2.VolumeHeader.FileName())
	fmt.Printf("Elevations: %d\n", ar2.Elevations())

	// spew.Dump(ar2.VolumeHeader)
}
