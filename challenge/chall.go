package challenge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	DepType      string   `yaml:"deployment_type,omitempty"`
	DepPort      int      `yaml:"deployment_port,omitempty"`
	CPU          int      `yaml:"cpu,omitempty"`
	Memory       int      `yaml:"mem,omitempty"`
	ChallDir     string
	ChallCache   ChallCache
	PrevCache    ChallCache
}

type Hint struct {
    Hint    string `yaml:"hint"`
    Cost    int    `yaml:"cost"`
	Visible bool   `yaml:"visible"`
}

type ExportStruct struct {
	CategoryQuery string   `json:"category_query"`
	ChallQuery    string   `json:"chall_query"`
	HintsQuery    string   `json:"hints_query"`
	DepMeta       DepMeta  `json:"deployment_metadata"`
	HintsChanged  bool     `json:"hints_changed"`
	ChallChanged  bool     `json:"chall_changed"`
	DockerChanged []string `json:"docker_changed"`
	ResChanged    []string `json:"res_changed"`
}

type DepMeta struct {
	DepType   string `json:"deployment_type"`
	DepPort   int    `json:"deployment_port"`
	Subdomain string `json:"subdomain"`
	CPU       int    `json:"cpu"`
	Memory    int    `json:"mem"`
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

func isValidJobName(name string) bool {
	regex := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	jobNameRegex := regexp.MustCompile(regex)
	return jobNameRegex.MatchString(name) && len(name) <= 253
}

func isValidChallDir(path string) bool {
	files, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range(files) {
		if !entry.IsDir() && entry.Name() == "chall.yaml" {
			dirName := filepath.Base(filepath.Clean(path))
			if !isValidJobName(dirName) {
				logger.LogMessage("WARN", fmt.Sprintf("ignoring directory %s since it does not follow RFC 1123", path), "Parser")
				return false
			}
			return true
		}
	}

	return false
}

func GetChalls(path string, noCache bool) []Challenge {
	var challs []Challenge

	if isValidChallDir(path) {
		chall := parseChallFile(filepath.Join(path, "chall.yaml"))
		chall.ChallDir = path
		err := chall.validate()
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
					err := chall.validate()
					if err != nil {
						logger.LogMessage("ERROR", fmt.Sprintf("error in challenge format: %v", err), "Validator")
					} else {
						err := chall.GenerateCache(noCache)
						if err != nil {
							logger.LogMessage("WARN", fmt.Sprintf("couldn't generate cache for chall %s: %v", path, err), "Main")
						}
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

func convertToSubdomain(input string) string {
	subdomain := strings.ToLower(input)

	re := regexp.MustCompile(`[^a-z0-9-]`)
	subdomain = re.ReplaceAllString(subdomain, "-")

	subdomain = strings.Trim(subdomain, "-")

	if len(subdomain) > 63 {
		subdomain = subdomain[:63]
	}

	if subdomain == "" {
		subdomain = "example"
	}

	return subdomain
}

func (c *Challenge) GetExportStruct() (expString string, err error) {
	exp := &ExportStruct{}

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

	exp.DepMeta.DepType = c.DepType
	exp.DepMeta.DepPort = c.DepPort

	if c.CPU != 0 {
		exp.DepMeta.CPU = c.CPU
	} else {
		exp.DepMeta.CPU = 15
	}

	if c.Memory != 0 {
		exp.DepMeta.Memory = c.Memory
	} else {
		exp.DepMeta.Memory = 32
	}

	exp.DepMeta.Subdomain = convertToSubdomain(c.ChallName)

	expjson, err := json.Marshal(exp)
	if err != nil {
		return "", fmt.Errorf("cannot marshal export data: %v", err)
	}

	expString = string(expjson)

	return
}

func (exp *ExportStruct) updateChanges(chall *Challenge) error {
	cache := chall.PrevCache

	newDockerHashes := chall.ChallCache.DockerHashs
	cachedDockerHashes := cache.DockerHashs
	dockersToBuild := []string{}

	for key, val := range(newDockerHashes) {
		cachedHash := cachedDockerHashes[key]
		if cachedHash != val {
			dockersToBuild = append(dockersToBuild, key)
		}
	}

	newResHashes := chall.ChallCache.ResHashs
	cachedResHashes := cache.ResHashs
	resToUpload := []string{}

	for key, val := range(newResHashes) {
		cachedHash := cachedResHashes[key]
		if cachedHash != val {
			resToUpload = append(resToUpload, key)
		}
	}

	if chall.ChallCache.HintsHash != cache.HintsHash {
		exp.HintsChanged = true
	}

	if chall.ChallCache.ChallHash != cache.ChallHash {
		exp.ChallChanged = true
	}

	exp.DockerChanged = dockersToBuild
	exp.ResChanged = resToUpload

	return nil
}

func (c *Challenge) validate() error {
	if c.ChallName == "" {
		return fmt.Errorf("%s: chall_name is required", c.ChallDir)
	}
	if c.Type == "" {
		return fmt.Errorf("%s: type is required", c.ChallDir)
	}
	if c.CategoryName == "" {
		return fmt.Errorf("%s: category_name is required", c.ChallDir)
	}
	if c.Prompt == "" {
		return fmt.Errorf("%s: prompt is required", c.ChallDir)
	}
	if c.Points == 0 {
		return fmt.Errorf("%s: points is required", c.ChallDir)
	}
	if c.Flag == "" {
		return fmt.Errorf("%s: flag is required", c.ChallDir)
	}
	if c.Author == "" {
		return fmt.Errorf("%s: author is required", c.ChallDir)
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
		file, err := os.Stat(filepath.Join(c.ChallDir, "Dockerfiles"))
		if err != nil {
			return fmt.Errorf("chall '%s' has type %s but Dockerfiles directory was not found", c.ChallDir, c.Type)
		}

		if !file.IsDir() {
			return fmt.Errorf("not a directory: Dockerfiles")
		}

		images, _ := os.ReadDir(filepath.Join(c.ChallDir, "Dockerfiles"))
		for _, image := range(images) {
			_, err := os.Stat(filepath.Join(c.ChallDir, "Dockerfiles", image.Name(), "Dockerfile"))
			if err != nil {
				return fmt.Errorf("file Dockerfile was not found in directory '%s'", filepath.Join(c.ChallDir, "Dockerfiles", image.Name()))
			}

		}
	}

	if c.Type == "dynamic" {
		if c.DepPort == 0 || c.DepType == "" {
			return fmt.Errorf("%s: challenge type was dynamic but the deployment type and port were not mentioned", c.ChallDir)
		}

		if c.DepType != "http" && c.DepType != "nc" && c.DepType != "ssh" {
			return fmt.Errorf("%s: deployment type can be one of ('http', 'ssh' or 'nc')", c.ChallDir)
		}
	}

	return nil
}