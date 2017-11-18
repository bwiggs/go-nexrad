package render

import (
	"image"
	"image/draw"
	"math"

	"github.com/Sirupsen/logrus"
	"github.com/bwiggs/go-nexrad/archive2"
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

	// spew.Dump(radials[0])
	// logrus.Infof("VCP: %d", radials[0].VolumeData.VolumeCoveragePatternNumber)

	// FIXME: could be off by .5 degree if the first measure is actually on the half for full resolution products
	azimuth := math.Floor(float64(radials[0].Header.AzimuthAngle))

	for i, radial := range radials {
		_ = i
		logrus.Infof("Starting Radial -- Azimuth: %f, Res: %d", azimuth, radial.Header.AzimuthResolutionSpacingCode)

		xc := width / 2
		yc := height / 2
		radiusX, radiusY := 2.0, 2.0
		startAngle := azimuth * (math.Pi / 180.0)                             /* angles are specified */
		angle := radial.Header.AzimuthResolutionSpacing() * (math.Pi / 180.0) /* clockwise in radians           */
		azimuth += radial.Header.AzimuthResolutionSpacing()
		if azimuth == 360.0 {
			azimuth = 0.0
		}
		gc.SetLineWidth(2)
		gc.SetLineCap(draw2d.ButtCap)

		for _, dbz := range radial.ReflectivityData.RefData() {
			radiusX += 1.0
			radiusY += 1.0
			gc.SetStrokeColor(dbzColor(dbz))
			gc.MoveTo(xc+math.Cos(startAngle)*radiusX, yc+math.Sin(startAngle)*radiusY)
			gc.ArcTo(xc, yc, radiusX, radiusY, startAngle, angle)
			gc.Stroke()
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
