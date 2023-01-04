package lic

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Exclusions is a map that holds license filenames that are to be ignored.
type Exclusions map[string]struct{}

// ExcludedEXT is a map that holds file extension names that are to be ingnored. (.go, .js, .cs, etc.).
type ExcludedEXT map[string]struct{}

// ExclusionsJson is the json file name for all excluded files.
const ExclusionsJson = "excludedfiles.json"

// ExludedEXTJson is the json file name for all excluded file extensions.
const ExcludedEXTJson = "excludedextensions.json"

// InitExlusions creates an Exclusions map using the data stored in ExcludedJson.
func InitExclusions(gopath string) (Exclusions, error) {
	var nilMap Exclusions
	excludedFile, ok := os.LookupEnv("DES_EXCL") // Use the DES_EXCL environment variable to change the path of the Exclusions file.
	if !ok {
		excludedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", ExclusionsJson)
	}
	excl := make(Exclusions)
	err := InitJsonConfigs(excludedFile, &excl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No exclusion config files leveraged")
			return make(Exclusions), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return excl, nil
}

// InitExcludedEXT creates an ExcludedEXT map using the info stored in the ExcludedEXT file.
func InitExcludedEXT(gopath string) (ExcludedEXT, error) {
	var nilMap ExcludedEXT
	excludedFile, ok := os.LookupEnv("DES_EXT") // Use the DES_EXT environment variable to change the path of the ExludedEXT file.
	if !ok {
		excludedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", ExcludedEXTJson)
	}
	ext := make(ExcludedEXT)
	err := InitJsonConfigs(excludedFile, &ext)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No extension exclusion config files leveraged")
			return make(ExcludedEXT), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return ext, nil
}
