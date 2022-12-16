package lic

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// UnknownLicense is the key for any licenses that we cannot find in our DefinedJson
const UnknownLicense = "Unknown License"

// NoLicense is the key for any repos that we cannot find a license file in
const NoLicense = "No License"

// The Scanner struct is the main object we use for our FileWalk and ScanPath. It holds all the paths
// maps and other things we need.
type Scanner struct {
	Gopath        string
	ModPath       string
	DstPath       string
	ProjectSum    string
	Licensecanned bool
	Exclusions    Exclusions
	Licenses      Licenses
	LicenseType   map[string][]string
}

// InitScanner creates a scanner object for scan path.
func InitScanner(gopath, modpath, dstpath string) (*Scanner, error) {

	excls, err := InitExclusions(gopath)
	if err != nil {
		return nil, err
	}

	licenses, err := InitLicense(gopath)
	if err != nil {
		return nil, err
	}

	return &Scanner{Gopath: gopath, ModPath: modpath, DstPath: dstpath, Licensecanned: false, Licenses: licenses, Exclusions: excls, LicenseType: make(map[string][]string)}, nil
}

// DependencyCheck first conforms the dependency string provided by ScanPath into the correct format for
// the ModPath. It then checks the ModPath to make sure the dependency exists. This is to
// prep the paths for a filewalk in ScanPath.
func (s *Scanner) DependencyCheck(d string) string {
	parts := strings.Split(d, "/go.mod")
	if len(parts) <= 1 {
		return ""
	}
	ver := regexp.MustCompile(`\s`)
	caps := regexp.MustCompile(`[A-Z]`)
	dep := ver.ReplaceAllString(parts[0], "@")
	dep = caps.ReplaceAllStringFunc(dep, func(s string) string { return "!" + strings.ToLower(s) })
	parts = strings.Split(dep, "/")
	path := filepath.Join(parts...)
	path = filepath.Join(s.ModPath, path)
	fstat, err := os.Stat(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ""
		}
		log.Printf("Problem with stat: %v, File checked: %v", err, path)
		return ""
	}
	if !fstat.IsDir() {
		log.Printf("Path was expected to be a directory: %v", path)
		return ""
	}
	return path
}

// ScanPath scans the path entered in and starts the process of copying and classifying license files into
// LicFolder.
func (s *Scanner) ScanPath() error {
	dependencies := strings.SplitAfterN(s.ProjectSum, "\n", -1)
	toScan := ""
	for _, d := range dependencies {
		d = strings.TrimSpace(d)
		toScan = s.DependencyCheck(d)
		if toScan == "" {
			continue
		}
		err := filepath.Walk(toScan, s.FileWalk)
		if err != nil {
			return err
		}
		if !s.Licensecanned {
			_, ok := s.LicenseType[NoLicense]
			if !ok {
				s.LicenseType[NoLicense] = []string{LicPathCleanup(toScan, false)}
			} else {
				s.LicenseType[NoLicense] = append(s.LicenseType[NoLicense], LicPathCleanup(toScan, false))
			}
		}
		s.Licensecanned = false
	}
	err := CreateLicTypesFile(*s)
	if err != nil {
		return err
	}
	log.Println("Scan completed")
	return nil
}

// FileWalk is a Walkfn for filepath.Walk in ScanPath. It walks through all folders and files in a repo and looks for anything
// relating to a license.
func (s *Scanner) FileWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error with file walk: %v", err)
	}
	if !IsLicenseFile(path) {
		return nil
	}
	s.Licensecanned = true
	_, ok := s.Exclusions[info.Name()]
	if ok {
		return nil
	}
	if info.IsDir() {
		return nil
	}
	bs, err := os.ReadFile(filepath.Join(path))
	if err != nil {
		return fmt.Errorf("unable to read file: %v", err)
	}
	licInfo := DefinitionFormat(string(bs))
	classified := false
	for license, def := range s.Licenses {
		if strings.Contains(licInfo, def) {
			_, ok = s.LicenseType[license]
			if !ok {
				s.LicenseType[license] = []string{LicPathCleanup(path, false)}
			} else {
				s.LicenseType[license] = append(s.LicenseType[license], LicPathCleanup(path, false))
			}
			classified = true
			break
		}
	}
	if !classified {
		_, ok = s.LicenseType[UnknownLicense]
		if !ok {
			s.LicenseType[UnknownLicense] = []string{LicPathCleanup(path, false)}
		} else {
			s.LicenseType[UnknownLicense] = append(s.LicenseType[UnknownLicense], LicPathCleanup(path, false))
		}
	}
	err = CreateLicFolder(path, s.DstPath, bs)
	if err != nil {
		return err
	}
	return nil
}
