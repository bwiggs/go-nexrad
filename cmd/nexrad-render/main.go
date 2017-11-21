package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"

	"golang.org/x/image/colornames"

	"github.com/Sirupsen/logrus"
	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "nexrad-render",
	Short: "nexrad-render will create radar images out of NEXRAD Level 2 (archive 2) files.",
	Run:   run,
}

var inputFile string
var outputFile string
var colorScheme string
var logLevel string
var imageSize int32

var colorSchemes map[string]func(float32) color.Color

func init() {
	cmd.PersistentFlags().StringVarP(&inputFile, "file", "f", "", "archive 2 file to process")
	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output radar image")
	cmd.PersistentFlags().StringVarP(&colorScheme, "color-scheme", "c", "noaa", "color scheme to use. noaa, scope, pink")
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "warn", "log level, debug, info, warn, error")
	cmd.PersistentFlags().Int32VarP(&imageSize, "size", "s", 1024, "size in pixel of the output image")

	colorSchemes = make(map[string]func(float32) color.Color)
	colorSchemes["noaa"] = dbzColorNOAA
	colorSchemes["scope"] = dbzColorScope
	colorSchemes["pink"] = dbzColor
	colorSchemes["clean-air"] = dbzColorCleanAirMode
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.SetLevel(lvl)

	f, err := os.Open(inputFile)
	defer f.Close()
	if err != nil {
		logrus.Error(err)
		return
	}

	ar2 := archive2.Extract(f)
	render(ar2.ElevationScans[1])
}

func render(radials []*archive2.Message31) {
	width := float64(imageSize)
	height := float64(imageSize)

	canvas := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	draw.Draw(canvas, canvas.Bounds(), image.Black, image.ZP, draw.Src)
	gc := draw2dimg.NewGraphicContext(canvas)

	xc := width / 2
	yc := height / 2
	pxPerKm := width / 2 / 460
	firstGatePx := float64(radials[0].ReflectivityData.DataMomentRange) / 1000 * pxPerKm
	gateIntervalKm := float64(radials[0].ReflectivityData.DataMomentRangeSampleInterval) / 1000
	gateWidthPx := gateIntervalKm * pxPerKm

	for _, radial := range radials {

		// round to the nearest rounded azimuth for the given resolution.
		// ex: for radial 20.5432, round to 20.5
		azimuthAngle := float64(radial.Header.AzimuthAngle)
		azimuthSpacing := radial.Header.AzimuthResolutionSpacing()
		azimuth := math.Floor(azimuthAngle)
		if math.Floor(azimuthAngle+azimuthSpacing) > azimuth {
			azimuth += azimuthSpacing
		}
		startAngle := azimuth * (math.Pi / 180.0)      /* angles are specified */
		endAngle := azimuthSpacing * (math.Pi / 180.0) /* clockwise in radians           */

		// start drawing gates from the start of the first gate
		distanceX, distanceY := firstGatePx, firstGatePx
		gc.SetLineWidth(gateWidthPx)

		for _, dbz := range radial.ReflectivityData.RefData() {
			if dbz > 5 {
				gc.SetStrokeColor(colorSchemes[colorScheme](dbz))
				gc.MoveTo(xc+math.Cos(startAngle)*distanceX, yc+math.Sin(startAngle)*distanceY)
				gc.ArcTo(xc, yc, distanceX, distanceY, startAngle, endAngle)
				gc.Stroke()
			}

			distanceX += gateWidthPx
			distanceY += gateWidthPx
			azimuth += radial.Header.AzimuthResolutionSpacing()
		}
	}

	// Save to file
	draw2dimg.SaveToPngFile(outputFile, canvas)
}

func dbzColor(dbz float32) color.Color {
	if dbz < 5.0 {
		return colornames.Black
	} else if dbz >= 5.0 && dbz < 10.0 {
		return color.RGBA{0x9C, 0x9C, 0x9C, 0xFF}
	} else if dbz >= 10.0 && dbz < 15.0 {
		return color.RGBA{0x76, 0x76, 0x76, 0xFF}
	} else if dbz >= 15.0 && dbz < 20.0 {
		return color.RGBA{0xFF, 0xAA, 0xAA, 0xFF}
	} else if dbz >= 20.0 && dbz < 25.0 {
		return color.RGBA{0xEE, 0x8C, 0x8C, 0xFF}
	} else if dbz >= 25.0 && dbz < 30.0 {
		return color.RGBA{0xC9, 0x70, 0x70, 0xFF}
	} else if dbz >= 30.0 && dbz < 35.0 {
		return color.RGBA{0x00, 0xFB, 0x90, 0xFF}
	} else if dbz >= 35.0 && dbz < 40.0 {
		return color.RGBA{0x00, 0xBB, 0x00, 0xFF}
	} else if dbz >= 40.0 && dbz < 45.0 {
		return color.RGBA{0xFF, 0xFF, 0x70, 0xFF}
	} else if dbz >= 45.0 && dbz < 50.0 {
		return color.RGBA{0xD0, 0xD0, 0x60, 0xFF}
	} else if dbz >= 50.0 && dbz < 55.0 {
		return color.RGBA{0xFF, 0x60, 0x60, 0xFF}
	} else if dbz >= 55.0 && dbz < 60.0 {
		return color.RGBA{0xDA, 0x00, 0x00, 0xFF}
	} else if dbz >= 60.0 && dbz < 65.0 {
		return color.RGBA{0xAE, 0x00, 0x00, 0xFF}
	} else if dbz >= 65.0 && dbz < 70.0 {
		return color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	} else if dbz >= 70.0 && dbz < 75.0 {
		return color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
	return color.RGBA{0xE7, 0x00, 0xFF, 0xFF}
}

func dbzColorCleanAirMode(dbz float32) color.Color {
	if dbz < -28.0 {
		return colornames.Black
	} else if dbz >= -28.0 && dbz < -24.0 {
		return color.RGBA{0x9C, 0x9C, 0x9C, 0xFF}
	} else if dbz >= -24.0 && dbz < -20.0 {
		return color.RGBA{0x76, 0x76, 0x76, 0xFF}
	} else if dbz >= -20.0 && dbz < -16.0 {
		return color.RGBA{0xFF, 0xAA, 0xAA, 0xFF}
	} else if dbz >= -16.0 && dbz < -12.0 {
		return color.RGBA{0xEE, 0x8C, 0x8C, 0xFF}
	} else if dbz >= -12.0 && dbz < -8.0 {
		return color.RGBA{0xC9, 0x70, 0x70, 0xFF}
	} else if dbz >= -8.0 && dbz < -4.0 {
		return color.RGBA{0x00, 0xFB, 0x90, 0xFF}
	} else if dbz >= -4.0 && dbz < 0.0 {
		return color.RGBA{0x00, 0xBB, 0x00, 0xFF}
	} else if dbz >= 0.0 && dbz < 4.0 {
		return color.RGBA{0xFF, 0xFF, 0x70, 0xFF}
	} else if dbz >= 4.0 && dbz < 8.0 {
		return color.RGBA{0xD0, 0xD0, 0x60, 0xFF}
	} else if dbz >= 8.0 && dbz < 12.0 {
		return color.RGBA{0xFF, 0x60, 0x60, 0xFF}
	} else if dbz >= 12.0 && dbz < 16.0 {
		return color.RGBA{0xDA, 0x00, 0x00, 0xFF}
	} else if dbz >= 16.0 && dbz < 20.0 {
		return color.RGBA{0xAE, 0x00, 0x00, 0xFF}
	} else if dbz >= 20.0 && dbz < 24.0 {
		return color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	} else if dbz >= 24.0 && dbz < 28.0 {
		return color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	}
	return color.RGBA{0xE7, 0x00, 0xFF, 0xFF}
}

func dbzColorNOAA(dbz float32) color.Color {
	if dbz < 5.0 {
		return color.RGBA{0x00, 0x00, 0x00, 0x00}
	} else if dbz >= 5.0 && dbz < 10.0 {
		return color.RGBA{0x40, 0xe8, 0xe3, 0xFF}
	} else if dbz >= 10.0 && dbz < 15.0 {
		// 26A4FA
		return color.RGBA{0x26, 0xa4, 0xfa, 0xFF}
	} else if dbz >= 15.0 && dbz < 20.0 {
		// 0030ED
		return color.RGBA{0x00, 0x30, 0xed, 0xFF}
	} else if dbz >= 20.0 && dbz < 25.0 {
		// 49FB3E
		return color.RGBA{0x49, 0xfb, 0x3e, 0xFF}
	} else if dbz >= 25.0 && dbz < 30.0 {
		// 36C22E
		return color.RGBA{0x36, 0xc2, 0x2e, 0xFF}
	} else if dbz >= 30.0 && dbz < 35.0 {
		// 278C1E
		return color.RGBA{0x27, 0x8c, 0x1e, 0xFF}
	} else if dbz >= 35.0 && dbz < 40.0 {
		// FEF543
		return color.RGBA{0xfe, 0xf5, 0x43, 0xFF}
	} else if dbz >= 40.0 && dbz < 45.0 {
		// EBB433
		return color.RGBA{0xeb, 0xb4, 0x33, 0xFF}
	} else if dbz >= 45.0 && dbz < 50.0 {
		// F6952E
		return color.RGBA{0xf6, 0x95, 0x2e, 0xFF}
	} else if dbz >= 50.0 && dbz < 55.0 {
		// F80A26
		return color.RGBA{0xf8, 0x0a, 0x26, 0xFF}
	} else if dbz >= 55.0 && dbz < 60.0 {
		// CB0516
		return color.RGBA{0xcb, 0x05, 0x16, 0xFF}
	} else if dbz >= 60.0 && dbz < 65.0 {
		// A90813
		return color.RGBA{0xa9, 0x08, 0x13, 0xFF}
	} else if dbz >= 65.0 && dbz < 70.0 {
		// EE34FA
		return color.RGBA{0xee, 0x34, 0xfa, 0xFF}
	} else if dbz >= 70.0 && dbz < 75.0 {
		return color.RGBA{0x91, 0x61, 0xc4, 0xFF}
	}
	return color.RGBA{0xff, 0xff, 0xFF, 0xFF}
}

func dbzColorScope(dbz float32) color.Color {
	if dbz < 5.0 {
		return colornames.Black
	} else if dbz >= 5.0 && dbz < 10.0 {
		return color.RGBA{0x02, 0x0d, 0x02, 0xFF}
	} else if dbz >= 10.0 && dbz < 15.0 {
		return color.RGBA{0x04, 0x23, 0x03, 0xFF}
	} else if dbz >= 15.0 && dbz < 20.0 {
		return color.RGBA{0x11, 0x52, 0x0d, 0xFF}
	} else if dbz >= 20.0 && dbz < 25.0 {
		return color.RGBA{0x33, 0xba, 0x2b, 0xFF}
	} else if dbz >= 25.0 && dbz < 30.0 {
		return color.RGBA{0x43, 0xeb, 0x39, 0xFF}
	} else if dbz >= 30.0 && dbz < 35.0 {
		return color.RGBA{0xff, 0xFB, 0x45, 0xFF}
	} else if dbz >= 35.0 && dbz < 40.0 {
		return color.RGBA{0xf5, 0xcb, 0x39, 0xFF}
	} else if dbz >= 40.0 && dbz < 45.0 {
		return color.RGBA{0xFb, 0xab, 0x32, 0xFF}
	} else if dbz >= 45.0 && dbz < 50.0 {
		return color.RGBA{0xfa, 0x83, 0x2a, 0xFF}
	} else if dbz >= 50.0 && dbz < 55.0 {
		return color.RGBA{0xbb, 0x03, 0x13, 0xFF}
	} else if dbz >= 55.0 && dbz < 60.0 {
		return color.RGBA{0xf7, 0x06, 0x1d, 0xFF}
	} else if dbz >= 60.0 && dbz < 65.0 {
		return color.RGBA{0xf9, 0x64, 0x69, 0xFF}
	} else if dbz >= 65.0 && dbz < 70.0 {
		return color.RGBA{0xfa, 0x97, 0xcc, 0xFF}
	} else if dbz >= 70.0 && dbz < 75.0 {
		return color.RGBA{0xf7, 0x34, 0xf9, 0xFF}
	}
	return color.RGBA{0xff, 0xff, 0xFF, 0xFF}
}
