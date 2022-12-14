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
	mod.ScanPath()
	/*m := make(map[string]string)
	m["BSD"] = `Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.`
	m["Apache"] = "Apache License"
	m["CDDL"] = `COMMON DEVELOPMENT AND DISTRIBUTION LICENSE (CDDL)`
	m["GNU"] = "GNU General Public License"
	m["Eclipse Public"] = "ECLIPSE PUBLIC LICENSE (“AGREEMENT”)"
	m["Mozilla Public"] = `1.1. “Contributor”
	means each individual or legal entity that creates, contributes to the creation of, or owns Covered Software.

	1.2. “Contributor Version”
	means the combination of the Contributions of others (if any) used by a Contributor and that particular Contributor’s Contribution.

	1.3. “Contribution”
	means Covered Software of a particular Contributor.

	1.4. “Covered Software”
	means Source Code Form to which the initial Contributor has attached the notice in Exhibit A, the Executable Form of such Source Code Form, and Modifications of such Source Code Form, in each case including portions thereof.`

	bs, _ := json.MarshalIndent(m, "", "   ")
	os.WriteFile(filepath.Join("Config", "definedlicenses.json"), bs, os.ModePerm)*/
	/*scanner, err := lic.InitScanner()
	if err != nil {
		fmt.Println(err)
		return
	}
	license, err := lic.InitLicense(*scanner)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(license)
	for k, v := range license {
		license[k] = lic.DefinitionFormat(v)
	}
	fmt.Println(license)
	*/
	/*
		fmt.Println(lic.IsLicenseFile("Projectlcense"))
		fmt.Println((lic.IsLicenseFile("ProjectLICENSE")))
		fmt.Println((lic.IsLicenseFile("Project.License")))
		fmt.Println((lic.IsLicenseFile("Project.lIcense")))
	*/
}
