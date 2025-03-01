package challenge

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Atish03/isolet-cli/logger"
	"gopkg.in/yaml.v2"
)

func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func calculateDirectoryMD5(dirPath string) (string, error) {
	hasher := md5.New()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileHash, err := calculateFileMD5(path)
		if err != nil {
			return err
		}

		hasher.Write([]byte(fileHash))
		hasher.Write([]byte(filepath.Base(filepath.Clean(path))))

		return nil
	})

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hashStruct(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to serialize struct: %w", err)
	}

	hash := md5.Sum(jsonBytes)

	return hex.EncodeToString(hash[:]), nil
}

func isCustomChall(chall Challenge) bool {
	file, err := os.Stat(filepath.Join(chall.ChallDir, "deployment.yaml"))
	if err != nil {
		return false
	}

	if !isValidDeployment(file) {
		return false
	}

	return true
}

func (chall *Challenge) handleCustomChall() error {
	if isCustomChall(*chall) {
		chall.CustomDeploy.Custom = true

		yamlStr, err := os.ReadFile(filepath.Join(chall.ChallDir, "deployment.yaml"))
		if err != nil {
			return fmt.Errorf("cannot read deloyment.yaml file: %v", err)
		}

		namespace := "isolet"

		if chall.Type == "dynamic" {
			namespace = "dynamic"
		}

		finalyamlStr := strings.ReplaceAll(string(yamlStr), "{{.Subd}}", ConvertToSubdomain(chall.ChallName))
		finalyamlStr = strings.ReplaceAll(finalyamlStr, "{{.Registry}}", filepath.Clean(chall.Registry.URL))
		finalyamlStr = strings.ReplaceAll(finalyamlStr, "{{.Namespace}}", namespace)

		chall.CustomDeploy.Deployment = finalyamlStr
	}

	return nil
}

func formatArray(arr []string) string {
	escapedElements := make([]string, len(arr))
	for i, v := range arr {
		escapedElements[i] = `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return "{" + strings.Join(escapedElements, ",") + "}"
}

func ConvertToSubdomain(input string) string {
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

func (exp *ExportStruct) updateChanges(chall *Challenge) {
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
}

func (exp *ExportStruct) populateResources(chall *Challenge) {
	exp.DepConfig.Resources.CPULimit = "30m"
	exp.DepConfig.Resources.MemLimit = "128Mi"

	exp.DepConfig.Resources.CPUReq = chall.CPU
	exp.DepConfig.Resources.MemReq = chall.Memory

	if exp.DepConfig.Resources.CPUReq == "" {
		exp.DepConfig.Resources.CPUReq = "10m"
	}

	if exp.DepConfig.Resources.MemReq == "" {
		exp.DepConfig.Resources.MemReq = "32Mi"
	}
}