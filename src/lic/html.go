package lic

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// InitLicTemplate creates the template used to convert all License files into .html files.
func InitLicTemplate() *template.Template {
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
	return template.Must(template.New("license").Parse(layout))

}

// CreateHtmlIndex creates an index .html file that is used for organizing all copied files into html.
// The html index also makes a link that can be used to go to the scanned License's current github repo a
// and it uses the gitapi to get the current repo's license in case it's changed.
func (l *Launch) CreateHtmlIndex() error {
	err := os.Remove(filepath.Join(l.Dst, l.Scanner.LicFolder, "index.html"))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error prepping html: %w", err)
		}
	}
	fileinfo, err := os.OpenFile(filepath.Join(l.Dst, l.Scanner.LicFolder, "index.html"), os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
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
			 	 {{if $val2.GitLicense}}
				  <p><a href={{$val2.Filepath|safe}}>{{$val2.Filename}}</a>   <a href={{$val2.GitLink|safe}}>Current Repo</a>  <strong>(github api: {{$val2.GitLicense}})</strong>
			     {{else if $val2.GitLink}}
			  <p><a href={{$val2.Filepath|safe}}>{{$val2.Filename}}</a>   <a href={{$val2.GitLink|safe}}>Current Repo</a>
			      {{else}}
			  <p><a href={{$val2.Filepath|safe}}>{{$val2.Filename}}</a></p>
			     {{end}}
			  {{end}}
		  {{end}}
		  </ol>
		</body>
		</html>`
	tmpl := template.Must(template.New("licenses").Funcs(funcMap).Parse(layout))
	err = tmpl.Execute(fileinfo, l.Scanner.LicenseType)
	if err != nil {
		return fmt.Errorf("error executing html: %w", err)
	}
	return nil
}

// CreateHTMLLicense copies each indivual license into html to be viewed by the index file.
func (s *Scanner) CreateHTMLLicense(licPath string, data []byte) error {
	licFolder := filepath.Join(s.DstPath, s.LicFolder, "Licenses")
	err := os.MkdirAll(licFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error making license folder directory: %w", err)
	}
	licNameExt := LicPathCleanup(filepath.Dir(licPath), true)

	dstFileName := filepath.Base(licPath) + licNameExt + ".html"

	dstFile := filepath.Join(licFolder, dstFileName)
	dFile, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error with dst file directory: %w", err)
	}
	defer dFile.Close()
	lines := strings.Split(string(data), "\n")

	err = s.Template.Execute(dFile, lines)
	if err != nil {
		return fmt.Errorf("error executing html: %w", err)
	}
	log.Println("HTML LICENSE COPIED!!!")
	return nil
}
