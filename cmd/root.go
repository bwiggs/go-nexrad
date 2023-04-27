package cmd

import (
	"os"

	"github.com/bwiggs/go-nexrad/internal/archive2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "go-nexrad",
	Short: "Process NEXRAD data.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		lvl, err := logrus.ParseLevel(logLevel)

		if err != nil {
			logrus.Fatalf("failed to parse level: %s", err)
		}

		logrus.SetLevel(lvl)
	},
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
