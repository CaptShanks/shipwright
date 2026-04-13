package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the ship CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ship %s\n", cliVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
