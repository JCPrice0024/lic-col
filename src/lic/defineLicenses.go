package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type Licenses map[string]string

const DefinedJson = "definedlicenses.json"

// add edit license func

func InitLicense(scanner Scanner) (Licenses, error) {
	var nilMap Licenses
	definedFile, ok := os.LookupEnv("DES_LIC")
	if !ok {
		definedFile = filepath.Join(scanner.Gopath, "src", "github.com", "JCPrice0024", "lic-col", "Config", DefinedJson)
	}
	_, err := os.Stat(definedFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No license config files leveraged")
			return make(Licenses), nil
		}
		return nilMap, fmt.Errorf("error checking file: %v", err)
	}
	lic, err := os.ReadFile(definedFile)
	if err != nil {
		return nilMap, fmt.Errorf("error reading file: %v", err)
	}
	lics := make(Licenses)
	err = json.Unmarshal(lic, &lics)
	if err != nil {
		return nilMap, fmt.Errorf("error unmarshaling: %v", err)
	}
	fmt.Println("License registered!")
	return lics, nil
}

func DefinitionFormat(definition string) string {
	defFormat := regexp.MustCompile(`\s+`)
	return defFormat.ReplaceAllString(definition, "")
}
