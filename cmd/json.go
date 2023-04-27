package cmd

import (
	"os"

	"github.com/bwiggs/go-nexrad/internal/geo"
	"github.com/bwiggs/go-nexrad/internal/geojson"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// jsonCmd represents the json command
var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "A brief description of your command",
	Run:   runJson,
	Args:  cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(jsonCmd)
}

func runJson(cmd *cobra.Command, args []string) {
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
