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
	ProjectSum    string
	Licensecanned bool
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
	modPath := filepath.Join(gopath, "pkg", "mod")
	sum, err := os.ReadFile(sumpath)
	if err != nil {
		return nil, fmt.Errorf("error opening mod file: %v", sum)
	}

	return &Scanner{Gopath: gopath, ModPath: modPath, ProjectSum: string(sum), Licensecanned: false, LicenseType: make(map[string][]string)}, nil
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
	//=strings.Contains()
	for _, d := range dependencies {
		d = strings.TrimSpace(d)
		toScan = s.DependencyCheck(d)
		if toScan == "" {
			continue
		}
		filepath.Walk(toScan, s.FileWalk)
		if !s.Licensecanned {
			_, ok := s.LicenseType[NoLicense]
			if !ok {
				s.LicenseType[NoLicense] = []string{toScan}
			} else {
				s.LicenseType[NoLicense] = append(s.LicenseType[NoLicense], toScan)
			}
		}
		s.Licensecanned = false
	}
	fmt.Println(s.LicenseType)
	return nil
}

func (s *Scanner) FileWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error with file walk: %v", err)
	}
	if !IsLicenseFile(info.Name()) {
		return nil
	}
	s.Licensecanned = true
	excls, err := InitExclusions(*s)
	if err != nil {
		return err
	}
	_, ok := excls[info.Name()]
	if ok {
		return nil
	}

	defLics, err := InitLicense(*s)
	if err != nil {
		return err
	}

	bs, err := os.ReadFile(filepath.Join(path))
	if err != nil {
		return fmt.Errorf("unable to read file: %v", err)
	}
	licInfo := DefinitionFormat(string(bs))
	licenseClass := s.LicenseType
	classified := false
	for license, def := range defLics {
		if strings.Contains(licInfo, def) {
			_, ok = licenseClass[license]
			if !ok {
				licenseClass[license] = []string{path}
			} else {
				licenseClass[license] = append(licenseClass[license], path)
			}
			classified = true
			break
		}
	}
	if !classified {
		_, ok = s.LicenseType[UnknownLicense]
		if !ok {
			licenseClass[UnknownLicense] = []string{path}
		} else {
			licenseClass[UnknownLicense] = append(licenseClass[UnknownLicense], path)
		}
	}
	return nil
}
