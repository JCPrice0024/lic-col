package lic

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLicTestRepoHtml(t *testing.T) {
	launcher := Launch{
		Repo:         "https://github.com/JCPrice0024/lic-testRepo5",
		Dst:          "UnitTest",
		Version:      "",
		CleanupMod:   true,
		CleanupClone: true,
		ToHTML:       true,
		GitCheck:     false,
	}
	err := launcher.createHtmlIndex()
	if err == nil {
		t.Fatal("Exepected err got nil")
	}
	err = launcher.LaunchProgram()
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED SCAN: %v", err)
	}
	expected := map[string][]licenseInfo{}
	exp, err := os.Open(filepath.Join("c:", string(filepath.Separator), "Users", "coold", "go", "src", "github.com", "JCPrice0024", "lic-col", "Config", "expectedresults.json"))
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO OPEN: %v", err)
	}
	dec := json.NewDecoder(exp)
	err = dec.Decode(&expected)
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO DECODE: %v", err)
	}
	for k := range launcher.Scanner.LicenseType {
		if !reflect.DeepEqual(launcher.Scanner.LicenseType[k], expected[k]) {
			t.Fatalf("EXPECTED: %v GOT: %v", expected[k], launcher.Scanner.LicenseType[k])
		}
	}
	err = os.RemoveAll("UnitTest")
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO REMOVE: %v", err)
	}
}

func TestLicTestRepo(t *testing.T) {
	launcher := Launch{
		Repo:         "https://github.com/JCPrice0024/lic-testRepo5",
		Dst:          "UnitTest",
		Version:      "2a83acbb36102d199d1b02ad71fb03a16aa172db",
		CleanupMod:   true,
		CleanupClone: false,
		ToHTML:       false,
		GitCheck:     false,
	}
	err := createLicTypesFile(launcher.Scanner)
	if err == nil {
		t.Fatal("Exepected err got nil")
	}
	err = launcher.LaunchProgram()
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED SCAN: %v", err)
	}
	expected := map[string][]licenseInfo{}
	exp, err := os.Open(filepath.Join("c:", string(filepath.Separator), "Users", "coold", "go", "src", "github.com", "JCPrice0024", "lic-col", "Config", "expectedresults.json"))
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO OPEN: %v", err)
	}
	dec := json.NewDecoder(exp)
	err = dec.Decode(&expected)
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO DECODE: %v", err)
	}
	for k := range launcher.Scanner.LicenseType {
		if !reflect.DeepEqual(launcher.Scanner.LicenseType[k], expected[k]) {
			t.Fatalf("EXPECTED: %v GOT: %v", expected[k], launcher.Scanner.LicenseType[k])
		}
	}
	err = os.RemoveAll("UnitTest")
	if err != nil {
		log.Println(err)
		t.Fatalf("FAILED TO REMOVE: %v", err)
	}
}

func TestConfigErrs(t *testing.T) {
	var err error
	_, err = initExcludedEXT("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	_, err = initExclusions("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	_, err = initInclusions("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	_, err = initLicense("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	_, err = initOverrides("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	_, err = createCache("")
	if err != nil {
		t.Fatalf("No file should not be an error: %v", err)
	}
	err = createCacheFile("", completedApiCheck{})
	if err == nil {
		t.Fatalf("File shouldn't exist: %v", err)
	}
}

func TestGithubApiPull(t *testing.T) {
	scan := Scanner{
		GitUser:           "JCPrice0024",
		GitToken:          "NOTATOKEN",
		CompletedApiCheck: make(completedApiCheck),
		GitLicense:        "",
	}
	err := scan.getGitLicense(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "JCPrice0024", "lic-testRepo5"))
	if err != nil {
		t.Fatalf("Failed to get Git License: %v", err)
	}

	err = scan.getGitLicense(filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "JCPrice0024", "lic-testRepo5", "src"))
	if err != nil {
		t.Fatalf("Failed to get Git License a 2nd time: %v", err)
	}

	gitApi := repo{
		Remaining: 50,
	}
	if gitApi.calcGitApiSleep() {
		t.Fatal("Expected to be near limit")
	}
	gitApi.Remaining = 1500
	if gitApi.calcGitApiSleep() {
		t.Fatal("Expected not to be near limit")
	}
}
