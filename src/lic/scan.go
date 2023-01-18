package lic

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// unknownLicense is the key for any licenses that we cannot find in our DefinedJson.
const unknownLicense = "Unknown License"

// noLicense is the key for any repos that we cannot find a license file in.
const noLicense = "No License"

// The Scanner struct is the main object we use for our FileWalk and ScanPath. It holds all the paths
// maps and other things we need.
type Scanner struct {
	Gopath            string
	ModPath           string
	DstPath           string
	ProjectSum        string
	LicFolder         string
	GitUser           string
	GitToken          string
	GitLicense        string // License for current repo from github.
	ToHTML            bool
	Licensecanned     bool
	Template          *template.Template
	CompletedApiCheck completedApiCheck
	Exclusions        exclusions
	ExcludedEXT       excludedEXT
	Inclusions        inclusions
	Override          overrides
	Licenses          licenses
	LicenseType       map[string][]licenseInfo
}

// licenseInfo is a struct that holds License information for use in making the LicTypesFile and all html files.
type licenseInfo struct {
	Filepath   string
	Filename   string
	GitLink    string
	GitLicense string
}

// initScanner creates a scanner object for scan path.
func initScanner(gopath, modpath, dstpath, gitUser, gitToken string, tohtml bool) (*Scanner, error) {

	excls, err := initExclusions(gopath)
	if err != nil {
		return nil, err
	}

	exc, err := initExcludedEXT(gopath)
	if err != nil {
		return nil, err
	}

	inc, err := initInclusions(gopath)
	if err != nil {
		return nil, err
	}

	licenses, err := InitLicense(gopath)
	if err != nil {
		return nil, err
	}

	ovr, err := initOverrides(gopath)
	if err != nil {
		return nil, err
	}
	api, err := createCache(gopath)
	if err != nil {
		return nil, err
	}
	tmpl := initLicTemplate()

	return &Scanner{
		Gopath:            gopath,
		ModPath:           modpath,
		DstPath:           dstpath,
		GitUser:           gitUser,
		GitToken:          gitToken,
		ToHTML:            tohtml,
		Licensecanned:     false,
		CompletedApiCheck: api,
		Template:          tmpl,
		Licenses:          licenses,
		ExcludedEXT:       exc,
		Exclusions:        excls,
		Inclusions:        inc,
		Override:          ovr,
		LicenseType:       make(map[string][]licenseInfo)}, nil
}

// dependencyCheck first conforms the dependency string provided by ScanPath into the correct format for
// the ModPath. It then checks the ModPath to make sure the dependency exists. This is to
// prep the paths for a filewalk in ScanPath.
func (s *Scanner) dependencyCheck(d string) string {
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
		log.Printf("Problem with stat: %s, File checked: %s", err, path)
		return ""
	}
	if !fstat.IsDir() {
		log.Printf("Path was expected to be a directory: %s", path)
		return ""
	}
	return path
}

// getGitParts takes in a path that has github.com in it. It then splits the
// path into the parts used by github. (example output: [github.com owner reponame]).
func getGitParts(path string) []string {
	if !strings.Contains(path, "github.com") {
		return []string{}
	}
	basepath := ""
	if !strings.Contains(path, filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")) {
		basepath = filepath.Join(os.Getenv("GOPATH"), "src")
	} else {
		basepath = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}
	fi, err := os.Stat(path)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	if strings.Contains(path, "!") {
		caps := regexp.MustCompile(`!\w`)
		path = caps.ReplaceAllStringFunc(path, func(s string) string {
			split := regexp.MustCompile(`!`)
			s = split.ReplaceAllString(s, "")
			return strings.ToUpper(s)
		})
	}
	p := path
	if !fi.IsDir() {
		p = filepath.Dir(path)
	}
	p = strings.ReplaceAll(p, basepath+string(filepath.Separator), "")
	cln := regexp.MustCompile(`@v(.*)`)
	p = cln.ReplaceAllString(p, "")
	return strings.Split(p, string(filepath.Separator))
}

// getLink gets a link from the github repo if available.
func getLink(path string) string {
	parts := getGitParts(path)
	if len(parts) < 3 {
		return ""
	}
	link := fmt.Sprintf("https://%s/%s/%s", parts[0], parts[1], parts[2])
	return link
}

// getGitLicense get's the git license from the githubapi.
func (s *Scanner) getGitLicense(path string) error {
	var err error
	parts := getGitParts(path)
	gitLic := repo{}
	var apiErr error
	if len(parts) >= 3 && s.GitUser != "" && s.GitToken != "" {
		lic, ok := s.CompletedApiCheck[fmt.Sprintf("%s/%s", parts[1], parts[2])]
		if !ok {
			gitLic, apiErr = getRepoInfo(parts[1], parts[2], s.GitUser, s.GitToken)
			reachLimit := gitLic.calcGitApiSleep()
			if reachLimit {
				s.GitUser = ""
				s.GitToken = ""
			}
			s.GitLicense = gitLic.License.Name
			s.CompletedApiCheck[fmt.Sprintf("%s/%s", parts[1], parts[2])] = s.GitLicense
		} else {
			s.GitLicense = lic
		}
	}
	if apiErr != nil {
		log.Printf("Problem getting license info: %s/%s %v", parts[1], parts[2], apiErr)
		log.Printf("If this is a private repo enter a username and personal acess token as Command Line Args")
	}
	if len(s.CompletedApiCheck)%10 == 0 {
		err = createCacheFile(s.Gopath, s.CompletedApiCheck)
		if err != nil {
			return err
		}
	}
	return nil
}

// ScanPath scans the path entered in and starts the process of copying and classifying license files into
// LicFolder. They will be .html files if the -tohtml Command Line Arg is used. In addition if
// a -git-token and -git-user is provided the program will check the api and provide the current github license.
// This will make the program wait a second everytime it is called.
func (s *Scanner) ScanPath() error {
	var err error
	dependencies := strings.SplitAfterN(s.ProjectSum, "\n", -1)
	toScan := ""
	for _, d := range dependencies {
		d = strings.TrimSpace(d)
		toScan = s.dependencyCheck(d)
		if toScan == "" {
			continue
		}
		err = s.getGitLicense(toScan)
		if err != nil {
			return err
		}
		err := filepath.Walk(toScan, s.fileWalk)
		if err != nil {
			return err
		}
		if !s.Licensecanned {
			_, ok := s.LicenseType[noLicense]
			licInfo := licenseInfo{Filename: licPathCleanup(toScan, false),
				Filepath:   filepath.Dir(toScan),
				GitLink:    getLink(toScan),
				GitLicense: s.GitLicense}
			if !ok {
				s.LicenseType[noLicense] = []licenseInfo{licInfo}
			} else {
				s.LicenseType[noLicense] = append(s.LicenseType[noLicense], licInfo)
			}
		}
		s.Licensecanned = false
		s.GitLicense = ""
	}
	err = createLicTypesFile(*s)
	if err != nil {
		return err
	}
	log.Println("Scan completed")
	return nil
}

// fileWalk is a Walkfn for filepath.Walk in ScanPath. It walks through all folders and files in a repo and looks for anything
// relating to a license.
func (s *Scanner) fileWalk(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return fmt.Errorf("error with file walk: %w", err)
	}
	if !isLicenseFile(path) {
		ovrPath := strings.Split(path, s.ModPath+string(filepath.Separator))
		if len(ovrPath) == 2 {
			_, ok := s.Override[ovrPath[1]]
			if ok {
				s.Licensecanned = true
				return s.scanOverride(path, ovrPath[1])
			}
		}
		_, ok := s.Inclusions[info.Name()]
		if !ok {
			return nil
		}
	}
	s.Licensecanned = true

	if s.checkExcluded(path, info.Name()) {
		return nil
	}

	if info.IsDir() {
		return nil
	}
	s.Exclusions[path] = struct{}{}
	bs, err := os.ReadFile(filepath.Join(path))
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	s.checkLicenses(bs, path)
	if s.ToHTML {
		err = s.createHTMLLicense(path, bs)
	} else {
		err = s.createLicFolder(path, bs)
	}
	if err != nil {
		return err
	}

	return nil
}

// checkExcluded checks the given path/filename to see if they need to be excluded.
func (s *Scanner) checkExcluded(path, filename string) bool {
	_, ok := s.Exclusions[path]
	if ok {
		return true
	}
	_, ok = s.Exclusions[filename]
	if ok {
		return true
	}
	_, ok = s.ExcludedEXT[strings.ToLower(filepath.Ext(filename))]
	return ok
}

// checkLicenses checks the given license to see if it matches any of our defined licenses.
func (s *Scanner) checkLicenses(bs []byte, path string) {
	licDef := DefinitionFormat(string(bs))
	classified := false
	licenseHTMLPath := filepath.Base(path) + licPathCleanup(filepath.Dir(path), true) + ".html"
	licInfo := licenseInfo{Filename: licPathCleanup(path, false),
		Filepath:   fmt.Sprintf("Licenses/%s", licenseHTMLPath),
		GitLink:    getLink(path),
		GitLicense: s.GitLicense}
	for _, def := range s.Licenses {
		matchesAll := TestLicense(licDef, def, false)
		if matchesAll {
			_, ok := s.LicenseType[def.Name]
			if !ok {
				s.LicenseType[def.Name] = []licenseInfo{licInfo}
			} else {
				s.LicenseType[def.Name] = append(s.LicenseType[def.Name], licInfo)
			}
			classified = true
			break
		}
	}
	if !classified {
		_, ok := s.LicenseType[unknownLicense]
		if !ok {
			s.LicenseType[unknownLicense] = []licenseInfo{licInfo}
		} else {
			s.LicenseType[unknownLicense] = append(s.LicenseType[unknownLicense], licInfo)
		}
	}
}

// scanOverride scans all overrided files and creates the copies for them. It will be a .html file if -tohtml is used
// in the Command Line Args.
func (s *Scanner) scanOverride(path, ovrPath string) error {
	licOvr := s.Override[ovrPath].License + " " + "OVERRIDE"
	ovrFile := filepath.Join(path, s.Override[ovrPath].Filename)
	htmlOvr := fmt.Sprintf("Licenses/%s", filepath.Base(s.Override[ovrPath].Filename)+licPathCleanup(filepath.Dir(ovrFile), true)+".html")
	licInfo := licenseInfo{
		Filename:   licPathCleanup(ovrFile, false),
		Filepath:   htmlOvr,
		GitLink:    getLink(path),
		GitLicense: s.GitLicense,
	}

	_, ok := s.LicenseType[licOvr]
	if !ok {
		s.LicenseType[licOvr] = []licenseInfo{licInfo}
	} else {
		s.LicenseType[licOvr] = append(s.LicenseType[licOvr], licInfo)
	}
	bs, err := os.ReadFile(ovrFile)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	if s.ToHTML {
		err = s.createHTMLLicense(ovrFile, bs)
	} else {
		err = s.createLicFolder(ovrFile, bs)
	}
	if err != nil {
		return err
	}
	return nil
}

// TestLicense is a function to help track down the differences between a license file and one of the defined
// licenses.
func TestLicense(licDef string, def definedLicense, debug bool) (matchesAll bool) {
	matchesAll = true
	for _, line := range def.Lines {
		if !strings.Contains(licDef, line) {
			matchesAll = false
			if debug {
				log.Printf("Inside test license line failed: line: %v license: %v ", line, def.Name)
			}
			break
		}
	}
	return
}
