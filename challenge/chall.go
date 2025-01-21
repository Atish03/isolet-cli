package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Atish03/isolet-cli/logger"
	"gopkg.in/yaml.v2"
)

type Challenge struct {
    ChallName    string   `yaml:"chall_name"`
    Type         string   `yaml:"type"`
    CategoryName string   `yaml:"category_name"`
    Prompt       string   `yaml:"prompt"`
    Points       int      `yaml:"points"`
    Requirements []string `yaml:"requirements"`
    Files        []string `yaml:"files"`
    Flag         string   `yaml:"flag"`
    Hints        []Hint   `yaml:"hints"`
    Author       string   `yaml:"author"`
    Visible      bool     `yaml:"visible,omitempty"`
    Tags         []string `yaml:"tags"`
    Links        []string `yaml:"links"`
	ChallDir     string
}

type Hint struct {
    Hint string  `yaml:"hint"`
    Cost int     `yaml:"cost"`
	Visible bool `yaml:"visible"`
}

type ExportStruct struct {
	ChallQuery string `json:"chall_query"`
	HintsQuery string `json:"hints_query"`
}

func parseChallFile(filename string) (chall Challenge) {
	file, err := os.Open(filename)
	if err != nil {
		logger.LogMessage("ERROR", "unable to read file chall.yaml", "Parser")
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&chall)
	if err != nil {
		logger.LogMessage("ERROR", fmt.Sprintf("cannot decode yaml: %v", err), "Parser")
	}

	return
}

func isValidChallDir(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range(files) {
		if !entry.IsDir() && entry.Name() == "chall.yaml" {
			return true
		}
	}

	return false
}

func GetChalls(path string) []Challenge {
	var challs []Challenge

	if isValidChallDir(path) {
		chall := parseChallFile(filepath.Join(path, "chall.yaml"))
		chall.ChallDir = path
		err := chall.validate()
		if err != nil {
			logger.LogMessage("ERROR", fmt.Sprintf("error in challenge format: %v", err), "Validator")
		} else {
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
					err := chall.validate()
					if err != nil {
						logger.LogMessage("ERROR", fmt.Sprintf("error in challenge format: %v", err), "Validator")
					} else {
						challs = append(challs, chall)
					}
				}
			}
		}
	}

	return challs
}

func formatArray(arr []string) string {
	escapedElements := make([]string, len(arr))
	for i, v := range arr {
		escapedElements[i] = `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(escapedElements, ",") + "}"
}

func (c *Challenge) GetExportStruct() (exp *ExportStruct, err error) {
	exp = &ExportStruct{}

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

	challQuery := fmt.Sprintf(`INSERT INTO challenges (chall_name, type, prompt, points, files, flag, author, visible, tags, links) VALUES ('%s', '%s', '%s', %d, '%s', '%s', '%s', %s, '%s', '%s')`,
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
		escapedHint := strings.ReplaceAll(hint.Hint, "'", `\'`)
		values[i] = fmt.Sprintf("(__CHALL_ID__, '%s', %d, %s)", escapedHint, hint.Cost, vis)
	}

	hintQuery += strings.Join(values, ", ")

	exp.ChallQuery = challQuery
	exp.HintsQuery = hintQuery

	return
}

func (c *Challenge) validate() error {
	if c.ChallName == "" {
		return fmt.Errorf("chall_name is required")
	}
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}
	if c.CategoryName == "" {
		return fmt.Errorf("category_name is required")
	}
	if c.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if c.Points == 0 {
		return fmt.Errorf("points is required")
	}
	if c.Flag == "" {
		return fmt.Errorf("flag is required")
	}
	if c.Author == "" {
		return fmt.Errorf("author is required")
	}
	if !c.Visible {
		c.Visible = false
	}

	resourceNotFound := []string{}

	for _, resource := range(c.Files) {
		if _, err := os.Stat(filepath.Join(c.ChallDir, "resources", resource)); err != nil {
			resourceNotFound = append(resourceNotFound, resource)
		}
	}

	if len(resourceNotFound) != 0 {
		return fmt.Errorf("following resources were not found: %s", strings.Join(resourceNotFound, ", "))
	}

	if c.Type != "static" {
		if _, err := os.Stat(filepath.Join(c.ChallDir, "Dockerfile")); err != nil {
			return fmt.Errorf("chall type is %s but Dockerfile not found", c.Type)
		}
	}

	return nil
}