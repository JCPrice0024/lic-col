package lic

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Inclusions is a map that holds non-license filenames that are to be included.
type Inclusions map[string]struct{}

// Overrides is a map that holds Repo names that don't have an actual license but should still be included.
type Overrides map[string]Override

// Override is a struct that holds a License type and a Filename of an Overrides entry.
// Both are necessary to make an Override.
type Override struct {
	License  string
	Filename string
}

// InclusionsJson is the file that holds all included filenames.
const InclusionsJson = "includedfiles.json"

// OverrideJson is the file that holds all required information for the Overrides.
const OverrideJson = "overridelicense.json"

// InitInclusions creates an Inclusions map using the data stored in IncludedJson.
func InitInclusions(gopath string) (Inclusions, error) {
	var nilMap Inclusions
	includedFile, ok := os.LookupEnv("DES_INCL")
	if !ok {
		includedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", InclusionsJson)
	}
	incl := make(Inclusions)
	err := InitJsonConfigs(includedFile, &incl)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No inclusion config files leveraged")
			return make(Inclusions), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return incl, nil
}

// InitOverrides creates an Overrides map using the data stored in OverrideJson.
func InitOverrides(gopath string) (Overrides, error) {
	var nilMap Overrides
	includedFile, ok := os.LookupEnv("DES_OVER")
	if !ok {
		includedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", OverrideJson)
	}
	ovr := make(Overrides)
	err := InitJsonConfigs(includedFile, &ovr)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No inclusion config files leveraged")
			return make(Overrides), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	return ovr, nil
}
