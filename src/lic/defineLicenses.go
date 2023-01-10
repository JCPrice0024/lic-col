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

// licenses is a map that is used to check known licenses in filewalk.
type licenses map[string]string

// definedJson is the name of the json file that holds all known licenses.
const definedJson = "definedlicenses.json"

// initLicense creates a Licenses map using the data stored in DefinedJson.
func initLicense(gopath string) (licenses, error) {
	var nilMap licenses
	definedFile, ok := os.LookupEnv("DES_LIC")
	if !ok {
		definedFile = filepath.Join(gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", definedJson)
	}
	lics := make(licenses)
	err := initJsonConfigs(definedFile, &lics)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No license config files leveraged")
			return make(licenses), nil
		}
		return nilMap, fmt.Errorf("error checking file: %w", err)
	}
	for license, def := range lics {
		lics[license] = definitionFormat(def)
	}
	return lics, nil
}

// isLicenseFile is a simple regex used to determine if a filename is a license file or not.
func isLicenseFile(path string) bool {
	licenseFile := regexp.MustCompile(`(?i)(.*)license(.*)`)
	return licenseFile.MatchString(path)
}

// definitionFormat is a simple regex used to format license definitions for comparison.
func definitionFormat(definition string) string {
	defFormat := regexp.MustCompile(`[^A-Za-z]+`)
	return strings.ToUpper(defFormat.ReplaceAllString(definition, ""))
}
