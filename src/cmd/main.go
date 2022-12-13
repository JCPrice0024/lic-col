package main

import "github.com/JCPrice0024/lic-col.git/src/lic"

func main() {
	/*mod, err := lic.InitScanner()
	if err != nil {
		fmt.Println(err)
		return
	}
	mod.ScanPath()*/
	/*bsd := `Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.`*/
	def := `ECLIPSE PUBLIC LICENSE (“AGREEMENT”)`
	lic.InitLicense("Eclipse Public license", def)
}
