package render

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/colornames"

	"github.com/Sirupsen/logrus"
	"github.com/bwiggs/nexrad/archive2"
	"github.com/davecgh/go-spew/spew"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func Draw(radials []*archive2.Message31) {
	width := float64(2048)
	height := float64(2048)
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	draw.Draw(dest, dest.Bounds(), image.Black, image.ZP, draw.Src)
	gc := draw2dimg.NewGraphicContext(dest)

	spew.Dump(radials[0])
	logrus.Infof("VCP: %d", radials[0].VolumeData.VolumeCoveragePatternNumber)
	for i, radial := range radials {
		_ = i
		logrus.Infof("Starting Radial -- Azimuth: %f, Res: %d", radial.Header.AzimuthAngle, radial.Header.AzimuthResolutionSpacing)

		xc := width / 2
		yc := height / 2
		radiusX, radiusY := 1.0, 1.0
		startAngle := float64(radial.Header.AzimuthAngle) * (math.Pi / 180.0)        /* angles are specified */
		angle := float64(radial.Header.AzimuthResolutionSpacing) * (math.Pi / 180.0) /* clockwise in radians           */
		// spew.Dump(startAngle, angle)
		gc.SetLineWidth(1)
		gc.SetLineCap(draw2d.SquareCap)

		for j, dbz := range radial.ReflectivityData.RefData() {
			radiusX += 1.0
			radiusY += 1.0
			// logrus.Debugf("dbz %f", dbz)
			gc.SetStrokeColor(dbzColorNOAA(dbz))
			gc.MoveTo(xc+math.Cos(startAngle)*radiusX, yc+math.Sin(startAngle)*radiusY)
			gc.ArcTo(xc, yc, radiusX, radiusY, startAngle, angle)
			gc.Stroke()
			_ = j
			// if j >= 10 {
			// 	break
			// }
		}
	}

	// gc.SetFontData(draw2d.FontData{Name: "Arial", Family: draw2d.FontFamilySans, Style: draw2d.FontStyleBold | draw2d.FontStyleItalic})
	// // Set the fill text color to black
	// gc.SetFillColor(image.White)
	// gc.SetFontSize(30)
	// // Display Hello World
	// gc.FillStringAt("Hello World", 8, 52)

	// Save to file
	draw2dimg.SaveToPngFile("reflectivity.png", dest)
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
