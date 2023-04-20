package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/bwiggs/go-nexrad/archive2"
	geojson "github.com/paulmach/go.geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	proj "github.com/twpayne/go-proj/v10"
)

var cmd = &cobra.Command{
	Use:   "nexrad-json [flags] file",
	Short: "nexrad-json generates GeoJSON from NEXRAD Level 2 (archive 2) data files.",
	Run:   run,
	Args:  cobra.MinimumNArgs(1),
}

var (
	outputFile string
	logLevel   string
	product    string
	elevation  int
	runners    int
)

var validProducts = map[string]struct{}{"ref": {}, "vel": {}, "sw": {}, "rho": {}}

func init() {
	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "radar.json", "output file")
	cmd.PersistentFlags().StringVarP(&product, "product", "p", "ref", "product to produce. ex: ref, vel, sw, rho")
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "warn", "log level, debug, info, warn, error")
	cmd.PersistentFlags().IntVarP(&runners, "threads", "t", runtime.NumCPU(), "threads")
	cmd.PersistentFlags().IntVarP(&elevation, "elevation", "e", 1, "1-15")
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run(cmd *cobra.Command, args []string) {
	inputFile := args[0]

	lvl, err := logrus.ParseLevel(logLevel)

	if err != nil {
		logrus.Fatalf("failed to parse level: %s", err)
	}

	logrus.SetLevel(lvl)

	if _, ok := validProducts[product]; !ok {
		logrus.Fatalf("invalid product %s", product)
	}

	f, err := os.Open(inputFile)

	if err != nil {
		logrus.Error(err)
		return
	}

	defer f.Close()

	ar2 := archive2.Extract(f)

	radials := ar2.ElevationScans[elevation]

	radarRelativeBins := make([]*Bin, 0)

	for _, radial := range radials {
		points := radialToRelativePoints(radial, product)

		radarRelativeBins = append(radarRelativeBins, points...)
	}

	radar_lat := radials[0].VolumeData.Lat
	radar_lon := radials[0].VolumeData.Long

	ltp := fmt.Sprintf("+proj=ortho +lat_0=%v +lon_0=%v +x_0=0 +y_0=0 +ellps=WGS84 +units=m +no_defs", radar_lat, radar_lon)

	geographic := "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs"

	ecef := "+proj=geocent +datum=WGS84 +units=m +no_defs +type=crs"

	ltpToEcef, err := proj.NewCRSToCRS(ltp, ecef, nil)

	if err != nil {
		logrus.Fatalln(err)
	}

	ecefToGeographic, err := proj.NewCRSToCRS(ecef, geographic, nil)

	if err != nil {
		logrus.Fatalln(err)
	}

	featureCollection := geojson.NewFeatureCollection()

	for _, relativeBin := range radarRelativeBins {
		geoBin := relativeBinToGeographicBin(ltpToEcef, ecefToGeographic, relativeBin)

		featureCollection.AddFeature(geoBin.ToPoly())
	}

	file, err := os.Create(outputFile)

	if err != nil {
		logrus.Fatalln(err)
	}

	defer file.Close()

	json, err := featureCollection.MarshalJSON()

	if err != nil {
		logrus.Fatalln(err)
	}

	file.Write(json)
}

func radialToRelativePoints(radial *archive2.Message31, product string) []*Bin {
	azimuth := radial.Header.AzimuthAngle
	elevation := radial.Header.ElevationAngle

	gates, err := radial.ScaledDataForProduct(product)

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

		if gate == archive2.MomentDataBelowThreshold || gate == archive2.MomentDataFolded || gate < 0 {
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

		bin := Bin{
			A:     point1,
			B:     point2,
			C:     point3,
			D:     point4,
			Value: gate,
		}

		radarRelativeBins = append(radarRelativeBins, &bin)

		r = r2
	}

	return radarRelativeBins
}

func relativeBinToGeographicBin(ltpToEcef *proj.PJ, ecefToGeographic *proj.PJ, relativeBin *Bin) *Bin {
	ecef, err := relativeBin.Forward(ltpToEcef)

	if err != nil {
		logrus.Fatalln(err)
	}

	geo, err := ecef.Forward(ecefToGeographic)

	if err != nil {
		logrus.Fatalln(err)
	}

	return geo
}
