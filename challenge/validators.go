package challenge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	if c.Attempts == 0 {
		c.Attempts = 500
	}

	if c.Type != "on-demand" && c.Type != "static" && c.Type != "dynamic" {
		return fmt.Errorf("%s: challenge type can be one of ('on-demand', 'dynamic', 'static')", c.ChallDir)
	}

	resourceNotFound := []string{}

	for _, resource := range(c.Files) {
		if _, err := os.Stat(filepath.Join(c.ChallDir, "resources", resource)); err != nil {
			resourceNotFound = append(resourceNotFound, resource)
		}
	}

	if len(c.Files) != 0 {
		extraResources := []string{}
		resInDir, err := os.ReadDir(filepath.Join(c.ChallDir, "resources"))
		if err != nil {
			return fmt.Errorf("%s: cannot find resources directory", c.ChallDir)
		}

		for _, e := range(resInDir) {
			if !slices.Contains(c.Files, e.Name()) {
				extraResources = append(extraResources, e.Name())
			}
		}

		if len(extraResources) != 0 {
			return fmt.Errorf("%s: some extra resources were found in resources directory, please add them to files or delete them: (%s)", c.ChallDir, strings.Join(extraResources, ", "))
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

	if c.Type != "static" {
		if c.DepType == "" {
			return fmt.Errorf("%s: challenge type was %s but the deployment type and/or port were not mentioned", c.ChallDir, c.Type)
		}

		if c.Type == "dynamic" && c.DepType != "http" && c.DepPort == 0 {
			return fmt.Errorf("%s: plesae mention port when the dynamic challenge is not http", c.ChallDir)
		}

		if c.DepType != "http" && c.DepType != "nc" && c.DepType != "ssh" {
			return fmt.Errorf("%s: deployment type can be one of ('http', 'ssh' or 'nc')", c.ChallDir)
		}
	} else {
		c.DepPort = 443
		c.DepType = "http"
	}

	return nil
}