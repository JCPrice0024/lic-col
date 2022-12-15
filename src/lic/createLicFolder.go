package lic

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LicTypesFile is the json file where a list of all known files and the
// dependencies that use them are stored.
const LicTypesFile = "licensetypes.json"

// LicFolder is the folder name where all copied licenses and the LicTypesFile go.
const LicFolder = "License_Folder"

// CreateLicTypesFile simply creates the LicTypesFile.
func CreateLicTypesFile(scanner Scanner) error {
	_, err := os.Stat(scanner.DstPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("licfolder was never made license scan failed: %v", err)
		}
		return err
	}
	bs, err := json.MarshalIndent(scanner.LicenseType, "", "   ")
	if err != nil {
		return fmt.Errorf("error marshaling licenseTypes: %v", err)
	}

	return os.WriteFile(filepath.Join(scanner.DstPath, LicFolder, LicTypesFile), bs, os.ModePerm)
}

// CreateLicFolder copies all License files into a Licenses folder found in the LicFolder.
func CreateLicFolder(licPath, dstPath string, data []byte) error {
	fmt.Println(dstPath)
	licFolder := filepath.Join(dstPath, LicFolder, "Licenses")
	err := os.MkdirAll(licFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error making license folder directory: %v", err)
	}
	licNameExt := LicPathCleanup(filepath.Dir(licPath), true)

	dstFileName := filepath.Base(licPath) + "_" + licNameExt

	dstFile := filepath.Join(licFolder, dstFileName)
	dFile, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error with dst file directory: %v", err)
	}
	defer dFile.Close()

	_, err = dFile.Write(data)
	if err != nil {
		return fmt.Errorf("error copying license file: %v", err)
	}
	log.Println("LICENSE COPIED!!!")
	return nil
}

// LicPathCleanup simply cleans the path up for CreateLicFolder and CreateLicTypesFile.
func LicPathCleanup(licPath string, noSlashes bool) string {
	lps := strings.Split(licPath, ("mod" + string(filepath.Separator)))
	if len(lps) != 2 {
		return ""
	}
	format := regexp.MustCompile(`/|\\`)
	if noSlashes {
		return format.ReplaceAllString(lps[1], "_")
	}

	return lps[1]
}
