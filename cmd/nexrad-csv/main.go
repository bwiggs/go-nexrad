package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/bwiggs/go-nexrad/archive2"
	proj "github.com/pebbe/proj/v5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "nexrad-csv [flags] file",
	Short: "nexrad-csv generates products from NEXRAD Level 2 (archive 2) data files.",
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

func init() {
	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "radar.csv", "output file")
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

type DataPoint struct {
	U     float64
	V     float64
	W     float64
	Value float32
}

func (p *DataPoint) ToRow() string {
	return fmt.Sprintf("%v,%v,%v\n", p.V, p.U, p.Value)
}

func run(cmd *cobra.Command, args []string) {

	inputFile := args[0]

	f, err := os.Open(inputFile)
	if err != nil {
		logrus.Error(err)
		return
	}
	defer f.Close()

	ar2 := archive2.Extract(f)
	logrus.Debug(ar2)

	radials := ar2.ElevationScans[elevation]

	radarRelativePoints := make([]*DataPoint, 0)

	for _, radial := range radials {
		azimuth := radial.Header.AzimuthAngle
		elevation := radial.Header.ElevationAngle
		gates := radial.ReflectivityData.ScaledData()

		firstGateDist := float64(radial.ReflectivityData.DataMomentRange)
		gateIncrement := float64(radial.ReflectivityData.DataMomentRangeSampleInterval)

		phi := 90 - elevation
		phi_radians := float64(phi * (math.Pi / 180))

		theta := 90 - azimuth

		if theta < 0 {
			theta += 360
		}

		theta_radians := float64(azimuth * (math.Pi / 180))

		r := firstGateDist

		for _, gate := range gates {

			if gate == archive2.MomentDataBelowThreshold {
				r += gateIncrement
				continue
			}

			point := &DataPoint{
				U:     r * math.Sin(phi_radians) * math.Cos(theta_radians),
				V:     r * math.Sin(phi_radians) * math.Sin(theta_radians),
				W:     r * math.Cos(phi_radians),
				Value: gate,
			}

			radarRelativePoints = append(radarRelativePoints, point)

			r += gateIncrement
		}
	}

	ctx := proj.NewContext()
	radar_lat := 39.78663
	radar_lon := -104.5458
	ltp := fmt.Sprintf("+proj=ortho +lat_0=%v +lon_0=%v +x_0=0 +y_0=0 +ellps=WGS84 +units=m +no_defs", radar_lat, radar_lon)
	geographic := "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs"
	ecef := "+proj=geocent +datum=WGS84 +units=m +no_defs +type=crs"

	ltpToEcef, err := ctx.CreateCRS2CRS(ltp, ecef)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ecefToGeographic, err := ctx.CreateCRS2CRS(ecef, geographic)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	geographicPoints := make([]*DataPoint, 0)

	// sample data
	for _, relativePoint := range radarRelativePoints {

		ecef_x, ecef_y, ecef_z, _, err := ltpToEcef.Trans(proj.Fwd, relativePoint.U, relativePoint.V, relativePoint.W, 0)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		lon, lat, alt, _, err := ecefToGeographic.Trans(proj.Fwd, ecef_x, ecef_y, ecef_z, 0)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		geoPoint := &DataPoint{
			U:     lon,
			V:     lat,
			W:     alt,
			Value: relativePoint.Value,
		}

		geographicPoints = append(geographicPoints, geoPoint)

	}

	file, err := os.Create(outputFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	// Header
	file.WriteString("latitude,longitude,reflectivity\n")

	for _, geoPoint := range geographicPoints {
		file.WriteString(geoPoint.ToRow())
	}
}
