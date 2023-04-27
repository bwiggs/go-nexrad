package cmd

import (
	"os"

	"github.com/jtleniger/go-nexrad-geojson/internal/archive2"
	"github.com/jtleniger/go-nexrad-geojson/internal/geo"
	"github.com/jtleniger/go-nexrad-geojson/internal/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "go-nexrad-json [NEXRAD archive file]",
	Short: "Create GeoJSON from NEXRAD data.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		lvl, err := logrus.ParseLevel(logLevel)

		if err != nil {
			logrus.Fatalf("failed to parse level: %s", err)
		}

		logrus.SetLevel(lvl)
	},
	Run:  run,
	Args: cobra.ExactArgs(1),
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
	archive2 := readArchive(args[0])

	bins := geo.RadarToBins(archive2, &geo.RadarToJSONOptions{
		Product: "ref",
		Minimum: 10,
	})

	o, err := os.Create("test.json")
	defer func() {
		o.Close()
	}()

	if err != nil {
		logrus.Fatal(err)
	}

	b := geojson.BinsToString(bins)

	o.WriteString(b.String())
}
