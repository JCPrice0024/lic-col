package lic

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Launch struct {
	Repo             string
	Dst              string
	Version          string
	Cleanup          bool
	Gopath           string
	ModPath          string
	CurrentDownloads map[string]struct{}
	Scanner          Scanner
}

func InitLaunch(repo, dst, version string, cleanup bool) (*Launch, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return nil, errors.New("no GOPATH found")
	}
	modpath := filepath.Join(gopath, "pkg", "mod")
	scan, err := InitScanner(gopath, modpath, dst)
	if err != nil {
		return nil, err
	}
	return &Launch{Repo: repo, Dst: dst, Version: version, Cleanup: cleanup, Gopath: gopath, ModPath: modpath, CurrentDownloads: make(map[string]struct{}), Scanner: *scan}, nil
}

func (l *Launch) LaunchProgram() error {
	var err error
	if l.Cleanup {
		filepath.Walk(l.ModPath, l.DownloadedWalk)
	}

	log.Println("Calling CloneRepo")
	clone, err := l.CloneRepo()
	if err != nil {
		return err
	}
	log.Println("CloneRepo completed")

	return filepath.Walk(clone, l.SumWalk)
}

func (l *Launch) CloneRepo() (string, error) {

	parts := strings.Split(l.Repo, "/")
	if len(parts) < 3 {
		return "", errors.New("invalid repo format")
	}

	path := []string{l.Gopath, "src"}
	path = append(path, parts[1:len(parts)-1]...)
	repoBase := filepath.Join(path...)
	repo := parts[len(parts)-1]
	extension := filepath.Ext(repo)
	repoDir := filepath.Join(repoBase, repo[0:len(repo)-len(extension)])

	log.Println("Calling Stat on: ", repoDir)
	_, err := os.Stat(repoDir)
	if err == nil {
		log.Println("repo already exists, analyzing current version.")
		return repoDir, nil
	}

	err = os.MkdirAll(repoBase, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error making directory: %v", err)
	}
	cloneRepo := exec.Command("git", "clone", l.Repo)
	cloneRepo.Dir = repoBase
	log.Println("Calling git clone")
	err = cloneRepo.Run()
	if err != nil {
		return "", fmt.Errorf("error running git clone command: %v", err)
	}
	log.Println("Clone completed")

	if l.Version != "" {
		gitCheckout := exec.Command("git", "checkout", l.Version)
		gitCheckout.Dir = repoDir
		log.Println("Calling git checkout")
		err = gitCheckout.Run()
		if err != nil {
			return "", fmt.Errorf("error running git checkout command: %v", err)
		}
		log.Println("Checkout completed")
	}
	return repoDir, nil
}

func (l *Launch) DownloadedWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error performing filepath.Walk: %v", err)
	}
	ver := regexp.MustCompile(`(.*)@v(.*)`)
	if !ver.MatchString(info.Name()) {
		return nil
	}
	_, ok := l.CurrentDownloads[path]
	if !ok {
		l.CurrentDownloads[path] = struct{}{}
	}
	return nil
}

func (l *Launch) SumWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error performing filepath.Walk: %v", err)
	}
	if info.Name() != "go.sum" {
		return nil
	}

	log.Println("Downloading sum data: ", path)
	goModDownload := exec.Command("go", "mod", "download")
	goModDownload.Dir = filepath.Dir(path)
	err = goModDownload.Run()
	if err != nil {
		return fmt.Errorf("error running go mod download files: %v", err)
	}
	log.Println("Download completed")
	sum, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error opening mod file: %v", sum)
	}
	l.Scanner.ProjectSum = string(sum)
	log.Println("Starting Scan")
	return l.Scanner.ScanPath()
}
