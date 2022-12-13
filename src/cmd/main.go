package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	/*mod, err := lic.InitScanner()
	if err != nil {
		fmt.Println(err)
		return
	}
	mod.ScanPath()*/
	m := make(map[string]string)
	m["BSD"] = `Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.`
	m["Apache"] = "Apache License"
	bs, _ := json.MarshalIndent(m, "", "   ")
	fmt.Println(string(bs))
}
