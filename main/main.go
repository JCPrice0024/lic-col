package main

import (
	"fmt"
	"os"
)

func main() {
	gopath, ok := os.LookupEnv("GOPATH")
	fmt.Println(gopath, ok)
	fmt.Println(licensescanner.ScanPath())
}
