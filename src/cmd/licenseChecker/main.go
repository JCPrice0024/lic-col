package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/JCPrice0024/lic-col/src/lic"
)

func main() {

	defLicenses, err := lic.InitLicense(os.Getenv("GOPATH"))
	if err != nil {
		log.Fatal(err)
	}

	fileToCheck := flag.String("filename", "", "the file you want to target")
	licToCheck := flag.String("license", "", "the license you want to check against")

	flag.Parse()

	if *fileToCheck == "" || *licToCheck == "" {
		flag.PrintDefaults()
		return
	}
	bs, err := os.ReadFile(*fileToCheck)
	if err != nil {
		log.Fatal(err)
	}

	license := lic.DefinitionFormat(string(bs))

	for _, v := range defLicenses {
		if strings.EqualFold(v.Name, *licToCheck) {
			lic.TestLicense(license, v, true)
		}
	}
}
