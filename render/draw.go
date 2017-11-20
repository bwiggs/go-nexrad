package render

import (
	"image"
	"image/draw"
	"math"

	"golang.org/x/image/colornames"

	"github.com/bwiggs/go-nexrad/archive2"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
)

func DrawSequential(radials []*archive2.Message31) {
	size := 1024
	width := float64(size)
	height := float64(size)
	pxPerKm := width / 2 / 460

	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	draw.Draw(dest, dest.Bounds(), image.Black, image.ZP, draw.Src)
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetLineWidth(1)
	gc.SetLineCap(draw2d.ButtCap)

	// // draw max radius
	// maxRange := 460 * pxPerKm
	// gc.SetStrokeColor(colornames.Darkgray)
	// gc.MoveTo(width/2+math.Cos(0)*maxRange, height/2+math.Sin(0)*maxRange)
	// gc.ArcTo(width/2, height/2, maxRange, maxRange, 0, 360)
	// gc.Stroke()

	// // draw half mac radius
	// visRange := 230 * pxPerKm
	// gc.SetStrokeColor(colornames.Darkgray)
	// gc.MoveTo(width/2+math.Cos(0)*visRange, height/2+math.Sin(0)*visRange)
	// gc.ArcTo(width/2, height/2, visRange, visRange, 0, 360)
	// gc.Stroke()

	// draw first gate range
	gc.SetStrokeColor(colornames.Darkgray)
	firstGatePx := float64(radials[0].ReflectivityData.DataMomentRange) / 1000 * pxPerKm
	gc.MoveTo(width/2+math.Cos(0)*firstGatePx, height/2+math.Sin(0)*firstGatePx)
	gc.ArcTo(width/2, height/2, firstGatePx, firstGatePx, 0, 360)
	gc.Stroke()

	draw2dimg.SaveToPngFile("reflectivity.spec.png", dest)

	gc.SetLineCap(draw2d.ButtCap)

	// logrus.Infof("VCP: %d", radials[0].VolumeData.VolumeCoveragePatternNumber)

	xc := width / 2
	yc := height / 2

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
				gc.SetStrokeColor(dbzColorNOAA(dbz))
				gc.MoveTo(xc+math.Cos(startAngle)*distanceX, yc+math.Sin(startAngle)*distanceY)
				gc.ArcTo(xc, yc, distanceX, distanceY, startAngle, endAngle)
				gc.Stroke()
			}

			distanceX += gateWidthPx
			distanceY += gateWidthPx
			azimuth += radial.Header.AzimuthResolutionSpacing()
		}
	}
	// gc.SetFontData(draw2d.FontData{Name: "Arial", Family: draw2d.FontFamilySans, Style: draw2d.FontStyleBold | draw2d.FontStyleItalic})
	// // Set the fill text color to black
	// gc.SetFillColor(image.White)
	// gc.SetFontSize(30)
	// // Display Hello World
	// gc.FillStringAt("Hello World", 8, 52)

	// Save to file
	draw2dimg.SaveToPngFile("reflectivity.spec.png", dest)
}
