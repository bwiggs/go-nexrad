package cmd

import (
	"os"
	"runtime/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read [NEXRAD archive]",
	Short: "Profiles reading a NEXRAD archive.",
	Run:   runProfileRead,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	profileCmd.AddCommand(readCmd)
}

func runProfileRead(cmd *cobra.Command, args []string) {
	f, err := os.Create("read.prof")

	if err != nil {
		logrus.Fatal(err)
	}

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	_ = readArchive(args[0])
}
