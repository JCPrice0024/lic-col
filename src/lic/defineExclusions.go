package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Exclusions is a map that holds license filenames that are to be ignored.
type Exclusions map[string]struct{}

// ExcludedJson is the json file name for all excludedfiles.
const ExcludedJson = "excludedfiles.json"

// InitExlusions creates an Exclusions map using the data stored in ExcludedJson.
func InitExclusions(gopath string) (Exclusions, error) {
	var nilMap Exclusions
	excludedFile, ok := os.LookupEnv("DES_EXCL")
	if !ok {
		excludedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", ExcludedJson)
	}
	_, err := os.Stat(excludedFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No exclusion config files leveraged")
			return make(Exclusions), nil
		}
		return nilMap, fmt.Errorf("error checking file: %v", err)
	}
	excl, err := os.ReadFile(excludedFile)
	if err != nil {
		return nilMap, fmt.Errorf("error reading file: %v", err)
	}
	excls := make(Exclusions)
	err = json.Unmarshal(excl, &excls)
	if err != nil {
		return nilMap, fmt.Errorf("error unmarshaling: %v", err)
	}
	return excls, nil
}
