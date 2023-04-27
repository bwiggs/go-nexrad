/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"runtime/pprof"

	"github.com/bwiggs/go-nexrad/internal/geo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// transformCmd represents the transform command
var transformCmd = &cobra.Command{
	Use:   "transform [NEXRAD archive]",
	Short: "Profiles transforming a NEXRAD archive.",
	Run:   runProfileTransform,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	profileCmd.AddCommand(transformCmd)
}

func runProfileTransform(cmd *cobra.Command, args []string) {
	f, err := os.Create("transform.prof")

	if err != nil {
		logrus.Fatal(err)
	}

	archive2 := readArchive(args[0])

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	bins := geo.RadarToBins(archive2, &geo.RadarToJSONOptions{
		Product: "ref",
	})

	logrus.Infof("%v", len(bins))
}
