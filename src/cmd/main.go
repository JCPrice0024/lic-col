package main

import (
	"fmt"

	"github.com/JCPrice0024/lic-col.git/src/lic"
)

func main() {
	mod, err := lic.InitScanner()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = mod.ScanPath()
	if err != nil {
		fmt.Println(err)
		return
	}
}
