package cmd

import (
	"os"
	"runtime/pprof"

	"github.com/jtleniger/go-nexrad-geojson/internal/geo"
	"github.com/jtleniger/go-nexrad-geojson/internal/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Create CPU profiles",
	Run:   runProfile,
}

func init() {
	rootCmd.AddCommand(profileCmd)
}

func runProfile(cmd *cobra.Command, args []string) {
	f, err := os.Create("profile.prof")

	if err != nil {
		logrus.Fatal(err)
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

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
