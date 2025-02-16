package cmd

import "github.com/spf13/cobra"

var registry string

func init() {
	initCmd.Flags().StringVar(&registry, "registry", "public", "Specify weather registry is private or public (default is public)")
}

var initCmd = &cobra.Command{
	Use: "config",
	Short: "configure cli to use the cluster",
	Long: "config command checks if the cluster has all the required resources and configure docker secret for private registries.",
	Run: func(cmd *cobra.Command, args []string) {
		configCLI(registry)
	},
}