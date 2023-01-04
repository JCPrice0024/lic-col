package lic

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Licenses is a map that is used to check known licenses in filewalk.
type Licenses map[string]string

// DefinedJson is the name of the json file that holds all known licenses.
const DefinedJson = "definedlicenses.json"

// InitLicense creates a Licenses map using the data stored in DefinedJson.
func InitLicense(gopath string) (Licenses, error) {
	var nilMap Licenses
	definedFile, ok := os.LookupEnv("DES_LIC")
	if !ok {
		definedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", DefinedJson)
	}
	lics := make(Licenses)
	err := InitJsonConfigs(definedFile, &lics)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No license config files leveraged")
			return make(Licenses), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	for license, def := range lics {
		lics[license] = DefinitionFormat(def)
	}
	return lics, nil
}

// IsLicenseFile is a simple regex used to determine if a filename is a license file or not.
func IsLicenseFile(path string) bool {
	licenseFile := regexp.MustCompile(`(?i)(.*)license(.*)`)
	return licenseFile.MatchString(path)
}

// DefinitionFormat is a simple regex used to format license definitions for comparison.
func DefinitionFormat(definition string) string {
	defFormat := regexp.MustCompile(`[^A-Za-z]+`)
	return strings.ToUpper(defFormat.ReplaceAllString(definition, ""))
}
