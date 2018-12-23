package cmd

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

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
