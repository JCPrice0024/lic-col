package lic

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const LicTypesFile = "licensetypes.json"
const LicFolder = "License_Folder"

func CreateLicTypesFile(scanner Scanner) error {
	bs, err := json.MarshalIndent(scanner.LicenseType, "", "   ")
	if err != nil {
		return fmt.Errorf("error marshaling licenseTypes: %v", err)
	}

	return os.WriteFile(filepath.Join(scanner.DstPath, LicFolder, LicTypesFile), bs, os.ModePerm)
}

func CreateLicFolder(licPath, dstPath string) error {
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

	lFile, err := os.Open(licPath)
	if err != nil {
		return fmt.Errorf("error opening license file: %v", err)
	}

	_, err = io.Copy(dFile, lFile)
	if err != nil {
		return fmt.Errorf("error copying license file: %v", err)
	}
	fmt.Println("LICENSE COPIED!!!")
	return nil
}

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
