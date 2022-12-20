package lic

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// func (s *Scanner) HandleHtml(writer http.ResponseWriter, request *http.Request) {
func (l *Launch) CreateHtmlIndex() error {
	fileinfo, err := os.OpenFile(filepath.Join(l.Dst, "index.html"), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer fileinfo.Close()
	//fmt.Println(l.Scanner.LicenseType)

	/*{{range $i, $val := .arr}}
	{{if lt $i 5}}<li>{{$val}}</li>{{end}}
	{{end}}*/
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
	/*link, err := createHtml(vals)
	if err != nil {
		fmt.Printf("error creating html links: %v", err)
	}
	*/
	err = tmpl.Execute(fileinfo, l.Scanner.LicenseType)
	if err != nil {
		return fmt.Errorf("error executing html: %v", err)
	}
	return nil
}

//func createHtml(vals []LicenseInfo)
