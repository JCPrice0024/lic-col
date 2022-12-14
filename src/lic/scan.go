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

const UnknownLicense = "Unknown License"
const NoLicense = "No License"

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

//type LicenseType map[string][]string

func InitScanner() (*Scanner, error) {
	sumpath := os.Getenv("PROJECTSUM")
	if sumpath == "" {
		return nil, errors.New("environment variable PROJECTSUM not set")
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil, errors.New("environment variable GOPATH not set")
	}
	dstpath := os.Getenv("DSTPATH")
	if dstpath == "" {
		return nil, errors.New("environment variable DSTPATH not set")
	}
	modPath := filepath.Join(gopath, "pkg", "mod")
	sum, err := os.ReadFile(sumpath)
	if err != nil {
		return nil, fmt.Errorf("error opening mod file: %v", sum)
	}

	excls, err := InitExclusions(gopath)
	if err != nil {
		return nil, err
	}

	licenses, err := InitLicense(gopath)
	if err != nil {
		return nil, err
	}

	return &Scanner{Gopath: gopath, ModPath: modPath, DstPath: dstpath, ProjectSum: string(sum), Licensecanned: false, Licenses: licenses, Exclusions: excls, LicenseType: make(map[string][]string)}, nil
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
	//fmt.Println("dependency: ", path)
	return path
}

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
	return nil
}

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
