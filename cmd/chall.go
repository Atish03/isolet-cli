package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Atish03/isolet-cli/challenge"
	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var kubecli = client.GetClient()

func init() {
	loadCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable cache while loading")
	deployCmd.Flags().BoolVar(&force, "force", false, "Force deploy a challenge")
	challCmd.AddCommand(lsCmd, testCmd, loadCmd, deployCmd, undeployCmd)
  	rootCmd.AddCommand(challCmd)
}

func drawChallTable(challs []challenge.Challenge) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Category", "Points", "Loaded on"})

	data := make([][]string, len(challs))

	for i, chall := range(challs) {
		timestamp := chall.PrevCache.TimeStamp.Format("Mon Jan 2 15:04:05")
		if chall.PrevCache.TimeStamp.IsZero() {
			timestamp = "-"
		}
		data[i] = []string {chall.ChallName, chall.Type, chall.CategoryName, fmt.Sprintf("%d", chall.Points), timestamp}
	}

	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func loadChalls(challs []challenge.Challenge) {
	var wg sync.WaitGroup
	
	for _, chall := range(challs) {
		wg.Add(1)
		
		go func(){
			err := chall.Load(&kubecli, "automation", "asia-south1-docker.pkg.dev/amiable-aquifer-449113-q1/pearlctf-dev/", &wg)
			if err != nil {
				logger.LogMessage("ERROR", fmt.Sprintf("error loading challenge: %v", err), "Main")
			}
		}()
	}

	wg.Wait()
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
		challs := challenge.GetChalls(challDir, false)
		drawChallTable(challs)
	},
}

var testCmd = &cobra.Command{
	Use:   "test <chall_name>",
	Short: "Test a specific challenge",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		challName := args[0]
		fmt.Printf("Testing challenge: %s\n", challName)
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
		challs := challenge.GetChalls(challDir, noCache)
		loadChalls(challs)
	},
}

var deployCmd = &cobra.Command{
	Use: "deploy <dir>",
	Short: "Deploy challenges/challenge in a directory\nNOTE: This will only search for dynamic challenges since static and on-demand don't require manual deployment",
	Run: func(cmd *cobra.Command, args []string) {
		dir := "."
		if len(args) != 0 {
			dir = args[0]
		}
		challDir, err := filepath.Abs(dir)
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("invalid directory %s", dir), "Main")
		}
		challs := challenge.GetChalls(challDir, noCache)
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
		challs := challenge.GetChalls(challDir, noCache)
		deleteChalls(challs)
	},
}