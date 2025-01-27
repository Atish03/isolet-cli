package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "isolet",
	Short: "isolet is a cli tool to manage ctf hosted on isolet platform",
	Long: `See isolet project at https://github.com/thealpha16/isolet`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("")
	},
  }
  
  func Execute() {
	if err := rootCmd.Execute(); err != nil {
	  fmt.Println(err)
	  os.Exit(1)
	}
  }