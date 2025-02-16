package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Atish03/isolet-cli/client"
	"github.com/Atish03/isolet-cli/logger"
)

func GetChalls(path string, noCache bool, cli *client.CustomClient) []Challenge {
	var challs []Challenge

	if isValidChallDir(path) {
		chall := parseChallFile(filepath.Join(path, "chall.yaml"))
		chall.ChallDir = path
		chall.Registry = cli.GetRegistry(chall.Type)

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
					chall.Registry = cli.GetRegistry(chall.Type)
					
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
	filesJSON := formatArray(c.Files)
	tagsJSON := formatArray(c.Tags)
	linksJSON := formatArray(c.Links)
	var vis string

	switch c.Visible {
	case true:
		vis = "TRUE"
	case false:
		vis = "FALSE"
	}

	categoryQuery := fmt.Sprintf(`
	INSERT INTO categories
	(category_name)
	VALUES ('%s')
	ON CONFLICT (category_name)
	DO UPDATE SET
		category_name = EXCLUDED.category_name
	RETURNING category_id
	`, c.CategoryName)

	challQuery := fmt.Sprintf(`
	INSERT INTO challenges
	(chall_name, category_id, type, prompt, points, files, flag, author, visible, tags, links)
	VALUES ('%s', $CATEGORY_ID, '%s', '%s', %d, '%s', '%s', '%s', %s, '%s', '%s')
	ON CONFLICT (chall_name)
	DO UPDATE SET
		category_id = EXCLUDED.category_id,
		type = EXCLUDED.type,
		prompt = EXCLUDED.prompt,
		points = EXCLUDED.points,
		files = EXCLUDED.files,
		author = EXCLUDED.author,
		visible = EXCLUDED.visible,
		tags = EXCLUDED.tags,
		links = EXCLUDED.links
	RETURNING chall_id
	`,
		c.ChallName,
		c.Type,
		c.Prompt,
		c.Points,
		string(filesJSON),
		c.Flag,
		c.Author,
		vis,
		string(tagsJSON),
		string(linksJSON),
	)

	hintQuery := `INSERT INTO hints (chall_id, hint, cost, visible) VALUES `
	values := make([]string, len(c.Hints))
	for i, hint := range c.Hints {
		switch hint.Visible {
		case true:
			vis = "TRUE"
		case false:
			vis = "FALSE"
		}
		escapedHint := strings.ReplaceAll(hint.Hint, "'", `''`)
		values[i] = fmt.Sprintf("($CHALL_ID, '%s', %d, %s)", escapedHint, hint.Cost, vis)
	}

	hintQuery += strings.Join(values, ", ")
	hintQuery += " RETURNING hid"

	exp.CategoryQuery = categoryQuery
	exp.ChallQuery = challQuery
	exp.HintsQuery = hintQuery

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