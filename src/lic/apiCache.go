package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// completedApiCheck is a map that holds all repos that have been checked using
// the githubapi.
type completedApiCheck map[string]string

// apiCacheJson is the file that holds all cached api Licenses.
const apiCacheJson = "cache.json"

// createCache creates a CompletedApiCheck map.
func createCache(gopath string) (completedApiCheck, error) {
	var nilMap completedApiCheck
	apiCache := filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", apiCacheJson)
	api := make(completedApiCheck)
	err := initJsonConfigs(apiCache, &api)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(completedApiCheck), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return api, nil
}

// createCacheFile creates a ApiCacheJson file, it stores License information about the repo gathered from the requests.
func createCacheFile(gopath string, cac completedApiCheck) error {
	apiCache := filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", apiCacheJson)
	bs, err := json.MarshalIndent(cac, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}
	return os.WriteFile(apiCache, bs, os.ModePerm)
}
