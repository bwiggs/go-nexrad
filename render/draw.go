package render

import (
	"image"
	"image/draw"
	"math"

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
