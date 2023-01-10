package lic

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// exclusions is a map that holds license filenames that are to be ignored.
type exclusions map[string]struct{}

// excludedEXT is a map that holds file extension names that are to be ingnored. (.go, .js, .cs, etc.).
type excludedEXT map[string]struct{}

// exclusionsJson is the json file name for all excluded files.
const exclusionsJson = "excludedfiles.json"

// excludedEXTJson is the json file name for all excluded file extensions.
const excludedEXTJson = "excludedextensions.json"

// InitExlusions creates an Exclusions map using the data stored in ExcludedJson.
func initExclusions(gopath string) (exclusions, error) {
	var nilMap exclusions
	excludedFile, ok := os.LookupEnv("DES_EXCL") // Use the DES_EXCL environment variable to change the path of the Exclusions file.
	if !ok {
		excludedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", exclusionsJson)
	}
	excl := make(exclusions)
	err := initJsonConfigs(excludedFile, &excl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No exclusion config files leveraged")
			return make(exclusions), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return excl, nil
}

// initExcludedEXT creates an ExcludedEXT map using the info stored in the ExcludedEXT file.
func initExcludedEXT(gopath string) (excludedEXT, error) {
	var nilMap excludedEXT
	excludedFile, ok := os.LookupEnv("DES_EXT") // Use the DES_EXT environment variable to change the path of the ExludedEXT file.
	if !ok {
		excludedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", excludedEXTJson)
	}
	ext := make(excludedEXT)
	err := initJsonConfigs(excludedFile, &ext)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No extension exclusion config files leveraged")
			return make(excludedEXT), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return ext, nil
}
