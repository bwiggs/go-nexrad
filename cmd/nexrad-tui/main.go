// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/jroimartin/gocui"
	"github.com/sirupsen/logrus"
)

var ar2 *archive2.Archive2
var currElevation = 3

const (
	ViewFileInfo         = "fileInfo"
	ViewElevationList    = "elevationList"
	ViewElevationDetails = "elevationDetails"
)

const logo = `░█▀█░█▀▀░█░█░█▀▄░█▀█░█▀▄
░█░█░█▀▀░▄▀▄░█▀▄░█▀█░█░█
░▀░▀░▀▀▀░▀░▀░▀░▀░▀░▀░▀▀░`

func loadElevationDetail(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()
	currElevation = cy + 1

	loadElevationData(g)
	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == ViewElevationList {
		_, err := g.SetCurrentView(ViewElevationDetails)
		return err
	}
	_, err := g.SetCurrentView(ViewElevationList)
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
		loadElevationDetail(g, v)
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
		loadElevationDetail(g, v)
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding(ViewElevationList, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(ViewElevationDetails, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding(ViewElevationList, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(ViewElevationList, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	// select new elevation
	if err := g.SetKeybinding(ViewElevationList, gocui.KeyEnter, gocui.ModNone, loadElevationDetail); err != nil {
		return err
	}

	if err := g.SetKeybinding(ViewElevationDetails, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding(ViewElevationDetails, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	fileInfoSize := 20

	leftPaneWidth := 30

	if v, err := g.SetView(ViewFileInfo, 0, 0, leftPaneWidth, fileInfoSize); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "File Info"
		fmt.Fprintf(v, "%s\n", logo)

		fmt.Fprintf(v, "File: %s\n", ar2.VolumeHeader.FileName())
		fmt.Fprintf(v, "ICAO %s\n", ar2.VolumeHeader.ICAO)
		fmt.Fprintf(v, "Build:%f\n", ar2.RadarStatus.GetBuildNumber())
		fmt.Fprintln(v, ar2.VolumeHeader.Date())
		fmt.Fprintf(v, "RDA Status: %s\n", ar2.RadarStatus.GetRDAStatus())
		fmt.Fprintf(v, "OP Status: %s\n", ar2.RadarStatus.GetOperabilityStatus())
		fmt.Fprintf(v, "VCP:%d\n", ar2.RadarStatus.VolumeCoveragePatternNum)
		fmt.Fprintf(v, "Alarms:%d\n", ar2.RadarStatus.AlarmCodes)
	}

	if v, err := g.SetView(ViewElevationList, 0, fileInfoSize, leftPaneWidth, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Elevations"
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		for _, e := range ar2.Elevations() {

			dat := []string{}
			first := ar2.ElevationScans[e][0]
			if first.ReflectivityData != nil {
				dat = append(dat, "REF")
			}
			if first.VelocityData != nil {
				dat = append(dat, "VEL")
			}
			if first.RhoData != nil {
				dat = append(dat, "RHO")
			}
			if first.SwData != nil {
				dat = append(dat, "SW")
			}
			if first.ZdrData != nil {
				dat = append(dat, "ZDR")
			}
			fmt.Fprintf(v, "%d %s\n", e, dat)
		}
	}

	if v, err := g.SetView(ViewElevationDetails, leftPaneWidth, 0, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Elevation Details"
		v.Wrap = true
		fmt.Fprintln(v, ar2.RadarPerformance)
		fmt.Fprintln(v, ar2.RadarStatus)
		if _, err := g.SetCurrentView(ViewElevationList); err != nil {
			return err
		}
	}
	return nil
}

func loadElevationData(g *gocui.Gui) {
	v, err := g.View(ViewElevationDetails)
	if err != nil {
		logrus.Fatalln(err)
	}
	v.Clear()
	for _, e := range ar2.ElevationScans[currElevation] {
		fmt.Fprintf(v, "ε:%f α:%f blkc:%d REF:%t VEL:%t\n", e.Header.ElevationAngle, e.Header.AzimuthAngle, e.Header.DataBlockCount, e.ReflectivityData != nil, e.VelocityData != nil)
	}
}

func main() {

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	f, err := os.Open(os.Args[1])
	logrus.SetLevel(logrus.InfoLevel)

	defer f.Close()
	if err != nil {
		logrus.Error(err)
		return
	}

	ar2 = archive2.Extract(f)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
