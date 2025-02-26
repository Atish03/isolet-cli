package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
)

func GetChalls(path string, noCache bool, cli *client.CustomClient) []Challenge {
	var challs []Challenge

	registry := cli.GetRegistry()

	if isValidChallDir(path) {
		chall := parseChallFile(filepath.Join(path, "chall.yaml"))
		chall.ChallDir = path
		chall.Registry = registry

		err := chall.handleCustomChall()
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("%v", err), "Main")
		}

		err = chall.validate()
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("error in challenge format: %v", err), "Validator")
		} else {
			err := chall.GenerateCache(noCache)
			if err != nil {
				logger.LogMessage("WARN", fmt.Sprintf("couldn't generate cache for chall %s: %v", path, err), "Main")
			}
			challs = append(challs, chall)
		}
	} else {
		all_chall_dirs, err := os.ReadDir(path)
		if err != nil {
			logger.LogMessage("ERROR", "Cannot read the specified directory", "Main")
			os.Exit(1)
		}

		for _, chall_dir := range(all_chall_dirs) {
			if chall_dir.IsDir() {
				full_chall_dir := filepath.Join(path, chall_dir.Name())
				if isValidChallDir(full_chall_dir) {
					chall := parseChallFile(filepath.Join(full_chall_dir, "chall.yaml"))
					chall.ChallDir = full_chall_dir
					chall.Registry = registry
					
					err := chall.handleCustomChall()
					if err != nil {
						logger.LogMessage("ERROR", fmt.Sprintf("%v", err), "Main")
					}

					err = chall.validate()
					if err != nil {
						logger.LogMessage("ERROR", fmt.Sprintf("error in challenge format: %v", err), "Validator")
						continue
					} else {
						err := chall.GenerateCache(noCache)
						if err != nil {
							logger.LogMessage("WARN", fmt.Sprintf("couldn't generate cache for chall %s: %v", path, err), "Main")
							continue
						}
						challs = append(challs, chall)
					}
				}
			}
		}
	}

	return challs
}

func (c *Challenge) GetExportStruct() (exp ExportStruct, err error) {
	tagsJSON := formatArray(c.Tags)
	linksJSON := formatArray(c.Links)
	var vis string

	switch c.Visible {
	case true:
		vis = "TRUE"
	case false:
		vis = "FALSE"
	}

	categoryValues := []string{c.CategoryName}

	challValues := []string {
		c.ChallName,
		c.Type,
		c.Prompt,
		strconv.Itoa(c.Points),
		c.Flag,
		c.Author,
		vis,
		string(tagsJSON),
		string(linksJSON),
		ConvertToSubdomain(c.ChallName),
		strconv.Itoa(c.DepPort),
		c.DepType,
		strconv.Itoa(c.Attempts),
	}

	hintValues := make([][]string, len(c.Hints))
	for i, hint := range c.Hints {
		switch hint.Visible {
		case true:
			vis = "TRUE"
		case false:
			vis = "FALSE"
		}
		hintValues[i] = []string{hint.Hint, strconv.Itoa(hint.Cost), vis}
	}

	exp.CategoryValues = categoryValues
	exp.ChallValues = challValues
	exp.HintsValues = hintValues

	exp.updateChanges(c)
	exp.populateResources(c)

	exp.DepConfig.DepType = c.DepType
	exp.DepConfig.DepPort = c.DepPort

	exp.DepConfig.Subdomain = ConvertToSubdomain(c.ChallName)

	exp.DepConfig.CustomDeploy = c.CustomDeploy
	exp.DepConfig.Registry = *c.Registry

	exp.OldName = c.PrevCache.ChallName
	if exp.OldName == "" {
		exp.OldName = c.ChallCache.ChallName
	}
	exp.NewName = c.ChallCache.ChallName

	return exp, nil
}