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
const DefinedFolder = "Defined_Licenses"

// add edit license func

func InitLicense(license string, definition string) error {
	if license == "" {
		log.Println("No license provided")
		return nil
	}
	if definition == "" {
		log.Println("No definition provided")
		return nil
	}
	definition = DefinitionFormat(definition)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	definedFile := filepath.Join(wd, DefinedFolder, DefinedJson)
	_, err = os.Stat(definedFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = CreateDefinedFile(license, definition)
			if err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("error checking file: %v", err)
	}
	lic, err := os.ReadFile(definedFile)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	lics := make(Licenses)
	err = json.Unmarshal(lic, &lics)
	if err != nil {
		return fmt.Errorf("error unmarshaling: %v", err)
	}
	if _, ok := lics[license]; ok {
		return nil
	}
	lics[license] = definition
	js, err := json.Marshal(lics)
	if err != nil {
		return fmt.Errorf("error marshaling file: %v", err)
	}
	err = os.WriteFile(definedFile, js, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}
	fmt.Println("License registered!")
	return nil
}

func CreateDefinedFile(license, definition string) error {
	definition = DefinitionFormat(definition)
	lic := Licenses{license: definition}
	js, err := json.Marshal(lic)
	if err != nil {
		return fmt.Errorf("error marshaling file: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}
	foldername := filepath.Join(wd, DefinedFolder)
	err = os.MkdirAll(foldername, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error making directory: %v", err)
	}
	filename := filepath.Join(foldername, DefinedJson)
	return os.WriteFile(filename, js, os.ModePerm)
}

func DefinitionFormat(definition string) string {
	defFormat := regexp.MustCompile(`\s+`)
	return defFormat.ReplaceAllString(definition, "")
}
