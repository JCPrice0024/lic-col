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

// definedLicense is the struct used to hold defined licenses.
type definedLicense struct {
	Name  string
	Lines []string
}

// licenses is a map that is used to check known licenses in filewalk.
type licenses []definedLicense

// definedJson is the name of the json file that holds all known licenses.
const definedJson = "definedlicenses.json"

// InitLicense creates a Licenses map using the data stored in DefinedJson.
func InitLicense(gopath string) (licenses, error) {
	var nilMap licenses
	definedFile, ok := os.LookupEnv("DES_LIC")
	if !ok {
		definedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", definedJson)
	}
	lics := make(licenses, 0)
	err := initJsonConfigs(definedFile, &lics)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No license config files leveraged")
			return make(licenses, 0), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	for licIndex, def := range lics {
		for lineIndex, line := range def.Lines {
			lics[licIndex].Lines[lineIndex] = DefinitionFormat(line)
		}
	}
	return lics, nil
}

// isLicenseFile is a simple regex used to determine if a filename is a license file or not.
func isLicenseFile(path string) bool {
	licenseFile := regexp.MustCompile(`(?i)(.*)license(.*)`)
	return licenseFile.MatchString(path)
}

// DefinitionFormat is a simple regex used to format license definitions for comparison.
func DefinitionFormat(definition string) string {
	defFormat := regexp.MustCompile(`[^A-Za-z]+`)
	return strings.ToUpper(defFormat.ReplaceAllString(definition, ""))
}
