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

// CreateLicTypesFile simply creates the LicTypesFile.
func CreateLicTypesFile(scanner Scanner) error {
	_, err := os.Stat(scanner.DstPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("licfolder was never made license scan failed: %w", err)
		}
		return err
	}
	bs, err := json.MarshalIndent(scanner.LicenseType, "", "   ")
	if err != nil {
		return fmt.Errorf("error marshaling licenseTypes: %w", err)
	}

	return os.WriteFile(filepath.Join(scanner.DstPath, scanner.LicFolder, LicTypesFile), bs, os.ModePerm)
}

// CreateLicFolder copies all License files into a Licenses folder found in the LicFolder.
func (s *Scanner) CreateLicFolder(licPath string, data []byte) error {
	licFolder := filepath.Join(s.DstPath, s.LicFolder, "Licenses")
	err := os.MkdirAll(licFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error making license folder directory: %w", err)
	}
	licNameExt := LicPathCleanup(filepath.Dir(licPath), true)

	dstFileName := filepath.Base(licPath) + licNameExt

	dstFile := filepath.Join(licFolder, dstFileName)
	dFile, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error with dst file directory: %w", err)
	}
	defer dFile.Close()

	_, err = dFile.Write(data)
	if err != nil {
		return fmt.Errorf("error copying license file: %w", err)
	}
	log.Println("LICENSE COPIED!!!")
	return nil
}

// LicPathCleanup simply cleans the path up for CreateLicFolder and CreateLicTypesFile.
func LicPathCleanup(licPath string, noSlashes bool) string {
	modpath := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	srcpath := filepath.Join(os.Getenv("GOPATH"), "src")
	var lps []string
	if !strings.Contains(licPath, modpath) {
		lps = strings.Split(licPath, srcpath)
	} else {
		lps = strings.Split(licPath, modpath)
	}
	if len(lps) == 1 {
		return ""
	}
	format := regexp.MustCompile(`/|\\`)
	if noSlashes {
		return format.ReplaceAllString(lps[1], "_")
	}

	return strings.TrimPrefix(lps[1], string(filepath.Separator))
}
