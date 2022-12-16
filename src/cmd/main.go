package main

import (
	"flag"
	"log"

	"github.com/JCPrice0024/lic-col.git/src/lic"
)

func main() {

	repo := flag.String("repo", "", "The repo flag is the github path you want to scan")
	dst := flag.String("dst", "", "The dst flag is the path where you want all of the scanned licenses to go")
	// clone := flag.String("clonepath", "", "The clonepath flag is the path of the clone of the repo you want scanned")
	cleanup := flag.Bool("cleanup", false, "The cleanup flag will clean up all downloaded folders once it's done")
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

	launcher, err := lic.InitLaunch(*repo, *dst, *version, *cleanup)
	if err != nil {
		log.Println(err)
		return
	}
	err = launcher.LaunchProgram()
	if err != nil {
		log.Println(err)
		return
	}
}
