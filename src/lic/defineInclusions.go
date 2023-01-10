package lic

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// inclusions is a map that holds non-license filenames that are to be included.
type inclusions map[string]struct{}

// overrides is a map that holds Repo names that don't have an actual license but should still be included.
type overrides map[string]override

// override is a struct that holds a License type and a Filename of an Overrides entry.
// Both are necessary to make an override.
type override struct {
	License  string
	Filename string
}

// inclusionsJson is the file that holds all included filenames.
const inclusionsJson = "includedfiles.json"

// overrideJson is the file that holds all required information for the Overrides.
const overrideJson = "overridelicense.json"

// initInclusions creates an Inclusions map using the data stored in IncludedJson.
func initInclusions(gopath string) (inclusions, error) {
	var nilMap inclusions
	includedFile, ok := os.LookupEnv("DES_INCL")
	if !ok {
		includedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", inclusionsJson)
	}
	incl := make(inclusions)
	err := initJsonConfigs(includedFile, &incl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No inclusion config files leveraged")
			return make(inclusions), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return incl, nil
}

// initOverrides creates an Overrides map using the data stored in OverrideJson.
func initOverrides(gopath string) (overrides, error) {
	var nilMap overrides
	includedFile, ok := os.LookupEnv("DES_OVER")
	if !ok {
		includedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", overrideJson)
	}
	ovr := make(overrides)
	err := initJsonConfigs(includedFile, &ovr)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No inclusion config files leveraged")
			return make(overrides), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return ovr, nil
}
