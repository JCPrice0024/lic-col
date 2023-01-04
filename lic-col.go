package main

import (
	"flag"
	"log"

	"github.com/JCPrice0024/lic-col/src/lic"
)

func main() {

	gitValidation := flag.Bool("git-check", false, "git-check allows for githubapi validation, it requires you to enter your github personal access token and username via Standard Input.")
	repo := flag.String("repo", "", "The repo flag is the github repo you'd like to scan, it can be in the form https://github.com/owner/reponame.git or git@github.com:owner/reponame.git")
	dst := flag.String("dst", "", "The dst flag is the path where you want all of the scanned licenses to go")
	cleanupMod := flag.Bool("clean-mod", false, "The clean-mod flag will remove all downloaded folders from the go mod download")
	cleanupClone := flag.Bool("clean-clone", false, "The clean-clone flag will remove all downloaded folders from the git clone")
	html := flag.Bool("tohtml", false, "tohtml copies all licenses into html, this makes the results of the scan much cleaner")
	version := flag.String("version", "", "The version flag is the commit hash of the repo you want to scan, if empty it scans the current version")

	flag.Parse()

	if *repo == "" {
		log.Println("No repo provided exiting")
		flag.PrintDefaults()
		return
	}
	if *dst == "" {
		log.Println("No dst provided exiting")
		flag.PrintDefaults()
		return
	}
	launcher := lic.Launch{
		Repo:         *repo,
		Dst:          *dst,
		Version:      *version,
		CleanupMod:   *cleanupMod,
		CleanupClone: *cleanupClone,
		ToHTML:       *html,
		GitCheck:     *gitValidation,
	}
	err := launcher.LaunchProgram()
	if err != nil {
		log.Println(err)
		return
	}
}
