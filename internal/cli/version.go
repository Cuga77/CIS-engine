package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the CLI version",
	Long:  `Prints the version number of the cis-cli binary.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cis-cli version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}