package geo

import (
	"math"

	"github.com/jtleniger/go-nexrad-geojson/internal/archive2"
	"github.com/sirupsen/logrus"
	"github.com/twpayne/go-proj/v10"
)

type RadarToJSONOptions struct {
	Product string
	Minimum float32
}

func RadarToBins(archive2 *archive2.Archive2, options *RadarToJSONOptions) []*Bin {
	volumeData := archive2.ElevationScans[1][0].VolumeData

	transforms := createTransforms(volumeData.Lat, volumeData.Lon)

	return georeferenceScan(archive2.ElevationScans[1], transforms, options)
}

func georeferenceScan(scan []*archive2.Message31, transforms []*proj.PJ, options *RadarToJSONOptions) []*Bin {
	bins := make([]*Bin, 0)

	for _, radial := range scan {
		relativeBins := radialToRelativePoints(radial, options)

		bins = append(bins, relativeBins...)
	}

	relativeBinsToGeographicBins(transforms, bins)

	return bins
}

func radialToRelativePoints(radial *archive2.Message31, options *RadarToJSONOptions) []*Bin {
	azimuth := radial.Header.AzimuthAngle
	elevation := radial.Header.ElevationAngle

	gates, err := radial.ScaledDataForProduct(options.Product)

	if err != nil {
		logrus.Fatalln(err)
	}

	firstGateDist := float64(radial.ReflectivityData.DataMomentRange)
	gateIncrement := float64(radial.ReflectivityData.DataMomentRangeSampleInterval)

	phi := 90 - elevation
	phi_radians := float64(phi * (math.Pi / 180))

	theta := 90 - azimuth

	if theta < 0 {
		theta += 360
	}

	thetaRadians := float64(theta * (math.Pi / 180))

	r := firstGateDist

	radarRelativeBins := make([]*Bin, 0)

	halfAzimuthSpacingRadians := radial.Header.AzimuthResolutionSpacing() * (math.Pi / 360)

	sinPhi := math.Sin(phi_radians)
	cosPhi := math.Cos(phi_radians)

	for _, gate := range *gates {
		r2 := r + gateIncrement

		// TODO: minimum value
		if gate == archive2.MomentDataBelowThreshold || gate == archive2.MomentDataFolded || gate < options.Minimum {
			r = r2
			continue
		}

		// From radar's point of view:
		// - bottom left
		// - bottom right
		// - top left
		// - top right
		point1 := proj.NewCoord(
			r*sinPhi*math.Cos(thetaRadians+halfAzimuthSpacingRadians),
			r*sinPhi*math.Sin(thetaRadians+halfAzimuthSpacingRadians),
			r*cosPhi,
			0,
		)

		point2 := proj.NewCoord(
			r*sinPhi*math.Cos(thetaRadians-halfAzimuthSpacingRadians),
			r*sinPhi*math.Sin(thetaRadians-halfAzimuthSpacingRadians),
			r*cosPhi,
			0,
		)

		point3 := proj.NewCoord(
			r2*sinPhi*math.Cos(thetaRadians+halfAzimuthSpacingRadians),
			r2*sinPhi*math.Sin(thetaRadians+halfAzimuthSpacingRadians),
			r2*cosPhi,
			0,
		)

		point4 := proj.NewCoord(
			r2*sinPhi*math.Cos(thetaRadians-halfAzimuthSpacingRadians),
			r2*sinPhi*math.Sin(thetaRadians-halfAzimuthSpacingRadians),
			r2*cosPhi,
			0,
		)

		bin := NewBin(point1, point2, point3, point4, gate)

		radarRelativeBins = append(radarRelativeBins, bin)

		r = r2
	}

	return radarRelativeBins
}

func relativeBinsToGeographicBins(transforms []*proj.PJ, relativeBins []*Bin) {
	allCoords := make([]proj.Coord, 0)

	for _, bin := range relativeBins {
		allCoords = append(allCoords, bin.Coords...)
	}

	for _, t := range transforms {
		t.ForwardArray(allCoords)
	}

	for i, bin := range relativeBins {
		bin.Coords = allCoords[(i * 4):(i*4 + 4)]
	}
}
