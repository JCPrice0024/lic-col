package licensescanner

import (
	"errors"
	"fmt"
	"os"
)

func ScanPath() error {
	modpath := os.Getenv("PROJECTMOD")
	if modpath == "" {
		return errors.New("environment variable PROJECTMOD not set")
	}
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("environment variable GOPATH not set")
	}
	mod, err := os.ReadFile(modpath)
	if err != nil {
		return fmt.Errorf("error opening mod file: %v", mod)
	}
	fmt.Println(string(mod))
	return nil
}

func CreateStore(licenseType string) {

}
