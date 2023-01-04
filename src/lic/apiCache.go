package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CompletedApiCheck is a map that holds all repos that have been checked using
// the githubapi.
type CompletedApiCheck map[string]string

// ApiCacheJson is the file that holds all cached api Licenses.
const ApiCacheJson = "cache.json"

// CreateCache creates a CompletedApiCheck map.
func CreateCache(gopath string) (CompletedApiCheck, error) {
	var nilMap CompletedApiCheck
	apiCache := filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", ApiCacheJson)
	api := make(CompletedApiCheck)
	err := InitJsonConfigs(apiCache, &api)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(CompletedApiCheck), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return api, nil
}

// CreateCacheFile creates a ApiCacheJson file, it stores License information about the repo gathered from the requests.
func CreateCacheFile(gopath string, cac CompletedApiCheck) error {
	apiCache := filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", ApiCacheJson)
	bs, err := json.MarshalIndent(cac, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}
	return os.WriteFile(apiCache, bs, os.ModePerm)
}
