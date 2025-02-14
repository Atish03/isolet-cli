package challenge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func (chall *Challenge) GenerateCache(noCache bool) error {
	if chall.Type != "static"{
		dirs, err := os.ReadDir(filepath.Join(chall.ChallDir, "Dockerfiles"))
		if err != nil {
			return fmt.Errorf("cannot read directory: %v", err)
		}

		dockerHashes := map[string]string{}

		for _, dir := range(dirs) {
			if dir.IsDir(){
				dirHash, err := calculateDirectoryMD5(filepath.Join(chall.ChallDir, "Dockerfiles", dir.Name()))
				if err != nil {
					return fmt.Errorf("cannot generate hash for %s: %v", dir.Name(), err)
				}
				dockerHashes[dir.Name()] = dirHash
			}
		}

		chall.ChallCache.DockerHashs = dockerHashes
	}

	if len(chall.Files) != 0 {
		dirs, err := os.ReadDir(filepath.Join(chall.ChallDir, "resources"))
		if err != nil {
			return fmt.Errorf("cannot read directory: %v", err)
		}

		resHashes := map[string]string{}

		for _, file := range(dirs) {
			if !file.IsDir(){
				fileHash, err := calculateFileMD5(filepath.Join(chall.ChallDir, "resources", file.Name()))
				if err != nil {
					return fmt.Errorf("cannot generate hash for %s: %v", file.Name(), err)
				}
				resHashes[file.Name()] = fileHash
			}
		}

		chall.ChallCache.ResHashs = resHashes
	}

	hintsHash, err := hashStruct(chall.Hints)
	if err != nil {
		return fmt.Errorf("cannot hash hints: %v", err)
	}

	challCopy := Challenge{
		ChallName: chall.ChallName,
		Type: chall.Type,
		CategoryName: chall.CategoryName,
		Prompt: chall.Prompt,
		Points: chall.Points,
		Requirements: chall.Requirements,
		Files: chall.Files,
		Flag: chall.Flag,
		Author: chall.Author,
		Visible: chall.Visible,
		Tags: chall.Tags,
		Links: chall.Links,
		DepType: chall.DepType,
		DepPort: chall.DepPort,
		CPU: chall.CPU,
		Memory: chall.Memory,
	}

	challHash, err := hashStruct(challCopy)
	if err != nil {
		return fmt.Errorf("cannot hash challenge: %v", err)
	}

	chall.ChallCache.HintsHash = hintsHash
	chall.ChallCache.ChallHash = challHash
	chall.ChallCache.ChallName = chall.ChallName

	prevCache := ChallCache{}

	if !noCache {
		cache_file := filepath.Join(chall.ChallDir, ".cache.json")
		if _, err := os.Stat(cache_file); err == nil {
			cacheContent, err := os.ReadFile(cache_file)
			if err != nil {
				return fmt.Errorf("cannot read cache file: %v", err)
			}
			err = json.Unmarshal(cacheContent, &prevCache)
			if err != nil {
				return fmt.Errorf("cannot unmarshal cache file: %v", err)
			}
		}
	}

	chall.PrevCache = prevCache

	return nil
}

func (chall *Challenge) SaveCache() error {
	chall.ChallCache.TimeStamp = time.Now()

	cacheJson, err := json.Marshal(chall.ChallCache)
	if err != nil {
		return fmt.Errorf("cannot marshal cache: %v", err)
	}

	fo, err := os.Create(filepath.Join(chall.ChallDir, ".cache.json"))
	if err != nil {
		return fmt.Errorf("cannot create cache: %v", err)
	}

	fo.Write(cacheJson)
	fo.Close()

	return nil
}