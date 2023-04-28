package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jtleniger/go-nexrad-geojson/internal/archive2"
	"github.com/jtleniger/go-nexrad-geojson/internal/geo"
	"github.com/jtleniger/go-nexrad-geojson/internal/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel       string
	minimum        float32
	maximum        float32
	product        string
	elevationRange string
	output         string
)

var validProducts = map[string]interface{}{"REF": "", "VEL": "", "SW": "", "ZDR": "", "PHI": "", "RHO": ""}

var rootCmd = &cobra.Command{
	Use:   "go-nexrad-json [NEXRAD archive file]",
	Short: "Create GeoJSON from NEXRAD data.",
	Run:   run,
	Args:  cobra.ExactArgs(1),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "warn", "set log level: debug, info, warn, error")
	rootCmd.PersistentFlags().Float32Var(&minimum, "minimum", 0, "minimum product value to include in the output")
	rootCmd.PersistentFlags().Float32Var(&maximum, "maximum", 0, "maximum prodct value to include in the output")
	rootCmd.PersistentFlags().StringVarP(&product, "product", "p", "REF", "product to output, one of REF, VEL, SW, ZDR, PHI, RHO, CFP")
	rootCmd.PersistentFlags().StringVarP(&elevationRange, "elevations", "e", "1", "elevation or range of elevations, can be N, or N-M (inclusive)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "radar", "base filename for output; elevation, product, and extension are appended")
}

func readArchive(filename string) *archive2.Archive2 {
	f, err := os.Open(filename)

	if err != nil {
		logrus.Fatal(err)
	}

	defer f.Close()

	return archive2.Extract(f)
}

func run(cmd *cobra.Command, args []string) {
	lvl, err := logrus.ParseLevel(logLevel)

	if err != nil {
		logrus.Fatalf("failed to parse level: %s", err)
	}

	logrus.SetLevel(lvl)

	opts := geo.RadarToJSONOptions{}

	if cmd.PersistentFlags().Changed("minimum") {
		opts.Minimum = &minimum
	}

	if cmd.PersistentFlags().Changed("maximum") {
		opts.Maximum = &maximum
	}

	product = strings.ToUpper(product)

	if _, ok := validProducts[product]; !ok {
		logrus.Fatalf("invalid product %v", product)
	}

	opts.Product = product

	elevationRegex, _ := regexp.Compile(`^(\d\d?|(\d\d?\-\d\d?))$`)

	if !elevationRegex.Match([]byte(elevationRange)) {
		logrus.Fatalf("invalid elevations %v", elevationRange)
	}

	elevations := strings.Split(elevationRange, "-")

	if len(elevations) == 1 {
		elevation, _ := strconv.Atoi(elevations[0])
		opts.Elevations = []int{elevation}
	} else {
		start, _ := strconv.Atoi(elevations[0])
		stop, _ := strconv.Atoi(elevations[1])

		if start >= stop {
			logrus.Fatalf("invalid elevations %v", elevationRange)
		}

		opts.Elevations = make([]int, 0)

		for i := start; i <= stop; i++ {
			opts.Elevations = append(opts.Elevations, i)
		}
	}

	archive2 := readArchive(args[0])

	bins := geo.RadarToBins(archive2, &opts)

	var wg sync.WaitGroup

	for elevation, scan := range bins {
		wg.Add(1)
		go func(elevation int, scan []*geo.Bin) {
			builder := geojson.BinsToString(scan)

			o, err := os.Create(fmt.Sprintf("%v-%v-%v.json", output, opts.Product, elevation))

			if err != nil {
				logrus.Fatal(err)
			}

			o.WriteString(builder.String())

			err = o.Close()

			if err != nil {
				logrus.Fatal(err)
			}

			wg.Done()
		}(elevation, scan)
	}

	wg.Wait()
}
