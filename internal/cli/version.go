package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию CLI",
	Long:  `Выводит номер версии бинарного файла cis-cli.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cis-cli version: %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
