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

type Scanner struct {
	ModPath    string
	ProjectSum string
}

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

	return &Scanner{ModPath: modPath, ProjectSum: string(sum)}, nil
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
		filepath.Walk(toScan, s.FileWalk)

	}
	return nil
}

func (s *Scanner) FileWalk(path string, info fs.FileInfo, err error) error {
	return nil
}
