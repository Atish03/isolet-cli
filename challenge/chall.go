package challenge

import (
	"fmt"
	"os"
	"path/filepath"

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
    Hint string `yaml:"hint"`
    Cost int    `yaml:"cost"`
}

func parseChallFile(filename string) Challenge {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Unable to read file chall.yaml")
		os.Exit(1)
	}
	defer file.Close()

	var chall Challenge
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&chall)
	if err != nil {
		fmt.Println("Cannot parse the yaml file:", filename)
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return chall
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
		err := chall.Validate()
		if err != nil {
			fmt.Printf("error in challenge format: %v", err)
		} else {
			challs = append(challs, chall)
		}
	} else {
		all_chall_dirs, err := os.ReadDir(path)
		if err != nil {
			fmt.Println("Cannot read the specified directory")
			os.Exit(1)
		}

		for _, chall_dir := range(all_chall_dirs) {
			if chall_dir.IsDir() {
				full_chall_dir := filepath.Join(path, chall_dir.Name())
				if isValidChallDir(full_chall_dir) {
					chall := parseChallFile(filepath.Join(full_chall_dir, "chall.yaml"))
					chall.ChallDir = full_chall_dir
					err := chall.Validate()
					if err != nil {
						fmt.Printf("error in challenge format: %s", err)
					} else {
						challs = append(challs, chall)
					}
				}
			}
		}
	}

	return challs
}

func (c *Challenge) Validate() error {
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
	return nil
}