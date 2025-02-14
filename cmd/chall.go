package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
	"github.com/spf13/cobra"
)

var kubecli = client.GetClient()

func init() {
	loadCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable cache while loading")
	deployCmd.Flags().BoolVar(&force, "force", false, "Force deploy a challenge")
	challCmd.AddCommand(lsCmd, loadCmd, deployCmd, undeployCmd)
  	rootCmd.AddCommand(challCmd)
}

var noCache bool
var force bool

var challCmd = &cobra.Command{
	Use:   "chall",
	Short: "manage challenges",
	Long: `
'chall' command is used to push, test, deploy challenges in k8s cluster

Challenges directory must follow a certain format
please refer https://github.com/Atish03/isolet-cli/# for more information on directory structure and chall.yaml.
	`,
}

var lsCmd = &cobra.Command{
	Use:   "ls <dir>",
	Short: "List all challenges in a particular directory",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}
		challDir, err := filepath.Abs(dir)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("invalid directory %s", dir), "Main")
		}
		challs := challenge.GetChalls(challDir, false, &kubecli)
		drawChallTable(challs)
	},
}

var loadCmd = &cobra.Command{
	Use:   "load <chall_name>",
	Short: "Load a specific challenge or challenges in directory",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}
		challDir, err := filepath.Abs(dir)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("invalid directory %s", dir), "Main")
		}
		challs := challenge.GetChalls(challDir, noCache, &kubecli)
		loadChalls(challs)
	},
}

var deployCmd = &cobra.Command{
	Use: "deploy <dir>",
	Short: "Deploy challenges/challenge in a directory",
	Long: "NOTE: This will only search for dynamic challenges since static and on-demand don't require manual deployment",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}
		challDir, err := filepath.Abs(dir)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("invalid directory %s", dir), "Main")
		}
		challs := challenge.GetChalls(challDir, noCache, &kubecli)
		deployChalls(challs, force)
	},
}

var undeployCmd = &cobra.Command{
	Use: "undeploy <dir>",
	Short: "Undeploy dynamic challenges",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}
		challDir, err := filepath.Abs(dir)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("invalid directory %s", dir), "Main")
		}
		challs := challenge.GetChalls(challDir, noCache, &kubecli)
		deleteChalls(challs)
	},
}