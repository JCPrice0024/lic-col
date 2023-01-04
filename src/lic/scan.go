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

// UnknownLicense is the key for any licenses that we cannot find in our DefinedJson.
const UnknownLicense = "Unknown License"

// NoLicense is the key for any repos that we cannot find a license file in.
const NoLicense = "No License"

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
	CompletedApiCheck CompletedApiCheck
	Exclusions        Exclusions
	ExcludedEXT       ExcludedEXT
	Inclusions        Inclusions
	Override          Overrides
	Licenses          Licenses
	LicenseType       map[string][]LicenseInfo
}

// LicenseInfo is a struct that holds License information for use in making the LicTypesFile and all html files.
type LicenseInfo struct {
	Filepath   string
	Filename   string
	GitLink    string
	GitLicense string
}

// InitScanner creates a scanner object for scan path.
func InitScanner(gopath, modpath, dstpath, gitUser, gitToken string, tohtml bool) (*Scanner, error) {

	excls, err := InitExclusions(gopath)
	if err != nil {
		return nil, err
	}

	exc, err := InitExcludedEXT(gopath)
	if err != nil {
		return nil, err
	}

	inc, err := InitInclusions(gopath)
	if err != nil {
		return nil, err
	}

	licenses, err := InitLicense(gopath)
	if err != nil {
		return nil, err
	}

	ovr, err := InitOverrides(gopath)
	if err != nil {
		return nil, err
	}
	api, err := CreateCache(gopath)
	if err != nil {
		return nil, err
	}
	tmpl := InitLicTemplate()

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
		LicenseType:       make(map[string][]LicenseInfo)}, nil
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
		log.Printf("Problem with stat: %s, File checked: %s", err, path)
		return ""
	}
	if !fstat.IsDir() {
		log.Printf("Path was expected to be a directory: %s", path)
		return ""
	}
	return path
}

// ScanOverride scans all overrided files and creates the copies for them. It will be a .html file if -tohtml is used
// in the Command Line Args.
func (s *Scanner) ScanOverride(path, ovrPath string) error {
	licOvr := s.Override[ovrPath].License + " " + "OVERRIDE"
	ovrFile := filepath.Join(path, s.Override[ovrPath].Filename)
	htmlOvr := fmt.Sprintf("Licenses/%s", filepath.Base(s.Override[ovrPath].Filename)+LicPathCleanup(filepath.Dir(ovrFile), true)+".html")
	_, ok := s.LicenseType[licOvr]
	if !ok {
		s.LicenseType[licOvr] = []LicenseInfo{{
			Filename:   LicPathCleanup(ovrFile, false),
			Filepath:   htmlOvr,
			GitLink:    GetLink(path),
			GitLicense: s.GitLicense}}
	} else {
		s.LicenseType[licOvr] = append(s.LicenseType[licOvr], LicenseInfo{
			Filename:   LicPathCleanup(ovrFile, false),
			Filepath:   htmlOvr,
			GitLink:    GetLink(path),
			GitLicense: s.GitLicense})
	}
	bs, err := os.ReadFile(ovrFile)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	if s.ToHTML {
		err = s.CreateHTMLLicense(ovrFile, bs)
	} else {
		err = s.CreateLicFolder(ovrFile, bs)
	}
	if err != nil {
		return err
	}
	return nil
}

// GetGitParts takes in a path that has github.com in it. It then splits the
// path into the parts used by github. (example output: [github.com owner reponame]).
func GetGitParts(path string) []string {
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

// GetLink gets a link from the github repo if available.
func GetLink(path string) string {
	parts := GetGitParts(path)
	if len(parts) < 3 {
		return ""
	}
	link := fmt.Sprintf("https://%s/%s/%s", parts[0], parts[1], parts[2])
	return link
}

// GetGitLicense get's the git license from the githubapi.
func (s *Scanner) GetGitLicense(path string) error {
	var err error
	parts := GetGitParts(path)
	gitLic := Repo{}
	var apiErr error
	if len(parts) >= 3 && s.GitUser != "" && s.GitToken != "" {
		lic, ok := s.CompletedApiCheck[fmt.Sprintf("%s/%s", parts[1], parts[2])]
		if !ok {
			gitLic, apiErr = GetRepoInfo(parts[1], parts[2], s.GitUser, s.GitToken)
			reachLimit := gitLic.CalcGitApiSleep()
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
		err = CreateCacheFile(s.Gopath, s.CompletedApiCheck)
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
		toScan = s.DependencyCheck(d)
		if toScan == "" {
			continue
		}
		err = s.GetGitLicense(toScan)
		if err != nil {
			return err
		}
		err := filepath.Walk(toScan, s.FileWalk)
		if err != nil {
			return err
		}
		if !s.Licensecanned {
			_, ok := s.LicenseType[NoLicense]
			licInfo := LicenseInfo{Filename: LicPathCleanup(toScan, false),
				Filepath:   filepath.Dir(toScan),
				GitLink:    GetLink(toScan),
				GitLicense: s.GitLicense}
			if !ok {
				s.LicenseType[NoLicense] = []LicenseInfo{licInfo}
			} else {
				s.LicenseType[NoLicense] = append(s.LicenseType[NoLicense], licInfo)
			}
		}
		s.Licensecanned = false
		s.GitLicense = ""
	}
	err = CreateLicTypesFile(*s)
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
		return fmt.Errorf("error with file walk: %w", err)
	}
	if !IsLicenseFile(path) {
		ovrPath := strings.Split(path, s.ModPath+string(filepath.Separator))
		if len(ovrPath) == 2 {
			_, ok := s.Override[ovrPath[1]]
			if ok {
				s.Licensecanned = true
				return s.ScanOverride(path, ovrPath[1])
			}
		}
		_, ok := s.Inclusions[info.Name()]
		if !ok {
			return nil
		}
	}
	s.Licensecanned = true

	s.checkExcluded(path, info.Name())

	if info.IsDir() {
		return nil
	}
	s.Exclusions[path] = struct{}{}
	bs, err := os.ReadFile(filepath.Join(path))
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	/*licDef := DefinitionFormat(string(bs))
	classified := false
	licenseHTMLPath := filepath.Base(path) + LicPathCleanup(filepath.Dir(path), true) + ".html"
	licInfo := LicenseInfo{Filename: LicPathCleanup(path, false),
		Filepath:   fmt.Sprintf("Licenses/%s", licenseHTMLPath),
		GitLink:    GetLink(path),
		GitLicense: s.GitLicense}
	for license, def := range s.Licenses {
		if strings.Contains(licDef, def) {
			_, ok = s.LicenseType[license]
			if !ok {
				s.LicenseType[license] = []LicenseInfo{licInfo}
			} else {
				s.LicenseType[license] = append(s.LicenseType[license], licInfo)
			}
			classified = true
			break
		}
	}
	if !classified {
		_, ok = s.LicenseType[UnknownLicense]
		if !ok {
			s.LicenseType[UnknownLicense] = []LicenseInfo{licInfo}
		} else {
			s.LicenseType[UnknownLicense] = append(s.LicenseType[UnknownLicense], licInfo)
		}
	}
	*/
	s.checkLicenses(bs, path)
	if s.ToHTML {
		err = s.CreateHTMLLicense(path, bs)
	} else {
		err = s.CreateLicFolder(path, bs)
	}
	if err != nil {
		return err
	}

	return nil
}

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

func (s *Scanner) checkLicenses(bs []byte, path string) {
	licDef := DefinitionFormat(string(bs))
	classified := false
	licenseHTMLPath := filepath.Base(path) + LicPathCleanup(filepath.Dir(path), true) + ".html"
	licInfo := LicenseInfo{Filename: LicPathCleanup(path, false),
		Filepath:   fmt.Sprintf("Licenses/%s", licenseHTMLPath),
		GitLink:    GetLink(path),
		GitLicense: s.GitLicense}
	for license, def := range s.Licenses {
		if strings.Contains(licDef, def) {
			_, ok := s.LicenseType[license]
			if !ok {
				s.LicenseType[license] = []LicenseInfo{licInfo}
			} else {
				s.LicenseType[license] = append(s.LicenseType[license], licInfo)
			}
			classified = true
			break
		}
	}
	if !classified {
		_, ok := s.LicenseType[UnknownLicense]
		if !ok {
			s.LicenseType[UnknownLicense] = []LicenseInfo{licInfo}
		} else {
			s.LicenseType[UnknownLicense] = append(s.LicenseType[UnknownLicense], licInfo)
		}
	}
}
