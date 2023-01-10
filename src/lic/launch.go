package lic

import (
	"bufio"
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

const filepathErrMsg = "error performing filepath.Walk: %w"

// Launch is a struct that holds all necessary info used to start the program and scan.
type Launch struct {
	Repo             string
	Dst              string
	Version          string
	CleanupMod       bool
	CleanupClone     bool
	ToHTML           bool
	GitCheck         bool
	Gopath           string
	ModPath          string
	CurrentDownloads map[string]struct{}
	Scanner          Scanner
}

// initLaunch creates a launch struct for use in starting the program.
func (l *Launch) initLaunch() error {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("no GOPATH found")
	}
	modpath := filepath.Join(gopath, "pkg", "mod")
	var token string
	var user string
	var err error
	if l.GitCheck {
		token, user, err = gitInfo()
		if err != nil {
			return err
		}
	}
	scan, err := initScanner(gopath, modpath, l.Dst, user, token, l.ToHTML)
	if err != nil {
		return err
	}
	l.Gopath = gopath
	l.ModPath = modpath
	l.CurrentDownloads = make(map[string]struct{})
	l.Scanner = *scan
	return nil
}

// LaunchProgram is the root of the program. It starts all processes declared by the
// Command Line Args.
func (l *Launch) LaunchProgram() error {
	var err error
	err = l.initLaunch()
	if err != nil {
		return err
	}
	if l.CleanupMod {
		filepath.Walk(l.ModPath, l.downloadedWalk)
	}

	log.Println("Calling CloneRepo")
	clone, err := l.cloneRepo()
	if err != nil {
		return err
	}

	l.Scanner.LicFolder = filepath.Base(clone) + "_" + "Licenses"

	log.Println("CloneRepo completed")

	log.Println("Scanning Cloned Repo")

	l.Scanner.getGitLicense(clone)

	err = filepath.Walk(clone, l.Scanner.fileWalk)
	if err != nil {
		return err
	}

	log.Println("Finished Scanning Cloned Repo")

	l.Scanner.GitLicense = ""
	err = filepath.Walk(clone, l.sumWalk)
	if err != nil {
		return err
	}
	if l.CleanupMod {
		log.Println("Cleaning mod path")
		err = filepath.Walk(l.ModPath, l.cleanerWalk)
		if err != nil {
			return err
		}
		log.Println("Cleaning Complete")
	}
	if l.CleanupClone {
		log.Println("Cleaning Clone")
		err = os.RemoveAll(clone)
		if err != nil {
			return fmt.Errorf("error removing clone: clone: %s err: %w", clone, err)
		}
		log.Println("Cleaning Complete")
	}

	if l.ToHTML {
		log.Println("Creating HTML Index")
		err = l.createHtmlIndex()
		if err != nil {
			return err
		}
	}
	log.Println("Exiting")
	return nil
}

// gitInfo safely gets the Git Username and Personal Access Token from Standard Input.
func gitInfo() (string, string, error) {
	var potentialUsername *bufio.Reader = bufio.NewReader(os.Stdin)
	log.Println("ENTER USERNAME: ")
	Username, err := potentialUsername.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("error reading input: %w", err)
	}
	var potentialToken *bufio.Reader = bufio.NewReader(os.Stdin)
	log.Println("ENTER PERSONAL ACCESS TOKEN: ")
	token, err := potentialToken.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("error reading input: %w", err)
	}
	token = strings.TrimSpace(token)
	return token, Username, nil
}

// cloneRepo performs a git clone on the provided repo. If there is a version tag
// cloneRepo also performs a git checkout on that version.
func (l *Launch) cloneRepo() (string, error) {
	repoDir := ""
	repoBase := ""
	if strings.Contains(l.Repo, "https://") {
		parts := strings.Split(l.Repo, "/")
		path := []string{l.Gopath, "src"}
		path = append(path, parts[1:len(parts)-1]...)
		repoBase = filepath.Join(path...)
		repo := parts[len(parts)-1]
		extension := filepath.Ext(repo)
		repoDir = filepath.Join(repoBase, repo[0:len(repo)-len(extension)])
	} else if strings.Contains(l.Repo, "git@") {
		ssh := strings.ReplaceAll(l.Repo, "git@", "")
		sshPath := strings.Replace(ssh, ":", "/", 1)
		parts := strings.Split(sshPath, "/")
		path := []string{l.Gopath, "src"}
		path = append(path, parts[:len(parts)-1]...)
		repoBase = filepath.Join(path...)
		repo := parts[len(parts)-1]
		extension := filepath.Ext(repo)
		repoDir = filepath.Join(repoBase, repo[0:len(repo)-len(extension)])
	}
	log.Println("Calling Stat on: ", repoDir)
	_, err := os.Stat(repoDir)
	if err == nil {
		log.Println("repo already exists, analyzing current version.")
		return repoDir, nil
	}
	err = os.MkdirAll(repoBase, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("error making directory: %w", err)
	}
	cloneRepo := exec.Command("git", "clone", l.Repo)
	cloneRepo.Dir = repoBase
	log.Println("Calling git clone")
	err = cloneRepo.Run()
	if err != nil {
		return "", fmt.Errorf("error running git clone command: %w", err)
	}
	log.Println("Clone completed")

	if l.Version != "" {
		gitCheckout := exec.Command("git", "checkout", l.Version)
		gitCheckout.Dir = repoDir
		log.Println("Calling git checkout")
		err = gitCheckout.Run()
		if err != nil {
			return "", fmt.Errorf("error running git checkout command: %w", err)
		}
		log.Println("Checkout completed")
	}
	return repoDir, nil
}

// cleanerWalk performs a filepath.Walk on the modpath at the end of the program and deletes all
// of the go mod downloads that were not there before. This is only called if -clean-mod is called.
func (l *Launch) cleanerWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf(filepathErrMsg, err)
	}
	ver := regexp.MustCompile(`(.*)@v(.*)`)
	if !ver.MatchString(info.Name()) {
		return nil
	}
	_, ok := l.CurrentDownloads[path]
	if ok {
		return nil
	}
	err = os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("error removing path:  path: %s err: %w", path, err)
	}
	return nil
}

// downloadedWalk is a helper function to CleanerWalk. It runs at the beginning of the program
// and performs a filepath.Walk on the modpath to prevent the deletion of any pre go mod downloads.
// This function is only called if you use -clean-mod.
func (l *Launch) downloadedWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf(filepathErrMsg, err)
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

// sumWalk performs a filepath.Walk on the provided repo and finds all go.sum files in the repo
// for scanning. It also performs the go mod download on the repo.
func (l *Launch) sumWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf(filepathErrMsg, err)
	}
	if info.Name() != "go.sum" {
		return nil
	}

	log.Println("Downloading sum data: ", path)
	goModDownload := exec.Command("go", "mod", "download")
	goModDownload.Dir = filepath.Dir(path)
	err = goModDownload.Run()
	if err != nil {
		return fmt.Errorf("error running go mod download files: %w", err)
	}
	log.Println("Download completed")
	sum, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error opening mod file: sum:%s err:%w", sum, err)
	}
	l.Scanner.ProjectSum = string(sum)
	log.Println("Starting Sum Scan")
	return l.Scanner.ScanPath()
}
