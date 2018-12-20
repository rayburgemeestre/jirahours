package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of jirahours",
	Long:  `All software has versions. This is jirahours's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The version is 1.0.0")
	},
}
