package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the lchj212 App version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
