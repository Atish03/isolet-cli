package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Atish03/isolet-cli/logger"
)

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

func isValidDeployment(file os.FileInfo) bool {
	// TODO: Add a validation for depoyment.yaml
	return true
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