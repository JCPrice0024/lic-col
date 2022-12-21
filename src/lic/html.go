package lic

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// func (s *Scanner) HandleHtml(writer http.ResponseWriter, request *http.Request) {
func (l *Launch) CreateHtmlIndex() error {
	fileinfo, err := os.OpenFile(filepath.Join(l.Dst, LicFolder, "index.html"), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer fileinfo.Close()

	funcMap := template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	layout := `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
		  <title>Scan Results</title>
		</head>
		<body>
		  <ol>
		  {{range $i, $val := .}}
		  <h1>{{$i}}</h1>
			  {{range $j, $val2 := $val}}
			  <p><a href={{$val2.Filepath|safe}}>{{$val2.Filename}}</a></p>
			  {{end}}
		  {{end}}
		  </ol>
		</body>
		</html>`
	tmpl := template.Must(template.New("licenses").Funcs(funcMap).Parse(layout))

	err = tmpl.Execute(fileinfo, l.Scanner.LicenseType)
	if err != nil {
		return fmt.Errorf("error executing html: %v", err)
	}
	return nil
}

// TODO: make html files for the licenses

func CreateHTMLLicense(licPath, dstPath string, data []byte) error {
	licFolder := filepath.Join(dstPath, LicFolder, "Licenses")
	err := os.MkdirAll(licFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error making license folder directory: %v", err)
	}
	licNameExt := LicPathCleanup(filepath.Dir(licPath), true)

	dstFileName := filepath.Base(licPath) + "_" + licNameExt + ".html"

	dstFile := filepath.Join(licFolder, dstFileName)
	dFile, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error with dst file directory: %v", err)
	}
	defer dFile.Close()
	lines := strings.Split(string(data), "\n")

	layout := `<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
		  <title>Scan Results</title>
		</head>
		<body>
		 <center>
		   {{range .}}
				<br/> {{.}}
		   {{end}}
		 </center>
		</body>
		</html>`
	tmpl := template.Must(template.New("license").Parse(layout))

	err = tmpl.Execute(dFile, lines)
	if err != nil {
		return fmt.Errorf("error executing html: %v", err)
	}
	log.Println("HTML LICENSE COPIED!!!")
	return nil
}
